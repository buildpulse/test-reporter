package submit

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awscreds "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/buildpulse/test-reporter/internal/logger"
	"github.com/buildpulse/test-reporter/internal/metadata"
	"github.com/buildpulse/test-reporter/internal/tar"
	"github.com/google/uuid"
)

type credentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

// A CommitResolverFactory provides methods for creating a
// metadata.CommitResolver.
type CommitResolverFactory interface {
	NewFromRepository(path string) (metadata.CommitResolver, error)
	NewFromStaticValue(commit *metadata.Commit) metadata.CommitResolver
}

type defaultCommitResolverFactory struct {
	logger logger.Logger
}

// NewCommitResolverFactory returns a CommitResolverFactory that creates
// CommitResolvers with the default production implementations.
func NewCommitResolverFactory(logger logger.Logger) CommitResolverFactory {
	return &defaultCommitResolverFactory{logger: logger}
}

// NewFromRepository returns a CommitResolver for looking up commits in the
// repository located at path.
func (d *defaultCommitResolverFactory) NewFromRepository(path string) (metadata.CommitResolver, error) {
	return metadata.NewRepositoryCommitResolver(path, d.logger)
}

// NewFromStaticValue returns a CommitResolver whose Lookup method always
// produces a Commit with values matching the fields in commit.
func (d *defaultCommitResolverFactory) NewFromStaticValue(commit *metadata.Commit) metadata.CommitResolver {
	return metadata.NewStaticCommitResolver(commit, d.logger)
}

// Submit represents the task of preparing and sending a set of test results to
// BuildPulse.
type Submit struct {
	client  *http.Client
	fs      *flag.FlagSet
	idgen   func() uuid.UUID
	logger  logger.Logger
	version *metadata.Version

	envs           map[string]string
	paths          []string
	bucket         string
	accountID      uint64
	endpointURL    string
	repositoryID   uint64
	repositoryPath string
	tree           string
	credentials    credentials
	commitResolver metadata.CommitResolver
}

// NewSubmit creates a new Submit instance.
func NewSubmit(version *metadata.Version, log logger.Logger) *Submit {
	s := &Submit{
		client:  http.DefaultClient,
		fs:      flag.NewFlagSet("submit", flag.ContinueOnError),
		idgen:   uuid.New,
		logger:  log,
		version: version,
	}

	s.fs.Uint64Var(&s.accountID, "account-id", 0, "BuildPulse account ID (required)")
	s.fs.Uint64Var(&s.repositoryID, "repository-id", 0, "BuildPulse repository ID (required)")
	s.fs.StringVar(&s.repositoryPath, "repository-dir", ".", "Path to local clone of repository")
	s.fs.StringVar(&s.endpointURL, "endpoint-url", "", "Hostname to point AWS client to")
	s.fs.StringVar(&s.tree, "tree", "", "SHA-1 hash of git tree")
	s.fs.SetOutput(ioutil.Discard) // Disable automatic writing to STDERR

	s.logger.Printf("Current version: %s", s.version.String())
	s.logger.Println("Initiating `submit`")

	return s
}

// Init populates s from args and envs. It returns an error if the required args
// or environment variables are missing or malformed.
func (s *Submit) Init(args []string, envs map[string]string, commitResolverFactory CommitResolverFactory) error {
	s.logger.Printf("Received args: %s", strings.Join(args, " "))

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	s.logger.Printf("Using working directory: %v", dir)

	pathArgs, flagArgs := pathsAndFlagsFromArgs(args)
	if len(pathArgs) == 0 {
		return fmt.Errorf("missing TEST_RESULTS_PATH")
	}

	s.paths, err = xmlPathsFromArgs(pathArgs)
	if err != nil {
		return err
	}
	if len(s.paths) == 0 {
		// To maintain backwards compatibility with releases prior to v0.19.0, if
		// exactly one path was given, and it's a directory, and it contains no XML
		// reports, continue without erroring. The resulting upload will contain
		// *zero* XML reports. In all other scenarios, treat this as an error.
		//
		// TODO: Treat this scenario as an error for the next major version release.
		info, err := os.Stat(pathArgs[0])
		isSingleDir := len(pathArgs) == 1 && err == nil && info.IsDir()
		if !isSingleDir {
			return fmt.Errorf("no XML reports found at TEST_RESULTS_PATH: %s", strings.Join(pathArgs, " "))
		}
	}

	if err := s.fs.Parse(flagArgs); err != nil {
		return err
	}

	flagset := make(map[string]bool)
	s.fs.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if s.accountID == 0 {
		return fmt.Errorf("missing required flag: -account-id")
	}

	if s.repositoryID == 0 {
		return fmt.Errorf("missing required flag: -repository-id")
	}

	if len(s.endpointURL) > 0 {
		if _, err := url.ParseRequestURI(s.endpointURL); err != nil {
			return fmt.Errorf("optional flag: -endpoint-url")
		}
	}

	id, ok := envs["BUILDPULSE_ACCESS_KEY_ID"]
	if !ok || id == "" {
		return fmt.Errorf("missing required environment variable: BUILDPULSE_ACCESS_KEY_ID")
	}
	s.credentials.AccessKeyID = id

	key, ok := envs["BUILDPULSE_SECRET_ACCESS_KEY"]
	if !ok || key == "" {
		return fmt.Errorf("missing required environment variable: BUILDPULSE_SECRET_ACCESS_KEY")
	}
	s.credentials.SecretAccessKey = key

	s.bucket, ok = envs["BUILDPULSE_BUCKET"]
	if !ok {
		s.bucket = "buildpulse-uploads"
	}

	if flagset["repository-dir"] && flagset["tree"] {
		return fmt.Errorf("invalid use of flag -repository-dir with flag -tree: use one or the other, but not both")
	}

	re := regexp.MustCompile(`^[0-9a-f]{40}$`)
	if flagset["tree"] && !re.MatchString(s.tree) {
		return fmt.Errorf("invalid value \"%s\" for flag -tree: should be a 40-character SHA-1 hash", s.tree)
	}

	info, err := os.Stat(s.repositoryPath)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid value for flag -repository-dir: %s is not a directory", s.repositoryPath)
	}

	s.envs = envs

	if flagset["tree"] {
		s.logger.Printf("Using value of -tree flag as the tree SHA for this submission: %s", s.tree)
		s.commitResolver = commitResolverFactory.NewFromStaticValue(&metadata.Commit{TreeSHA: s.tree})
		return nil
	}

	if !flagset["repository-dir"] {
		s.logger.Printf("Using default value for -repository-dir flag: %s", s.repositoryPath)
	}

	s.logger.Printf("Looking for git repository at %s", s.repositoryPath)
	s.commitResolver, err = commitResolverFactory.NewFromRepository(s.repositoryPath)
	if err != nil {
		return fmt.Errorf("invalid value for flag -repository-dir: %v", err)
	}
	s.logger.Printf("Found git repository at %s", s.repositoryPath)

	return nil
}

// Run packages up the test results and sends them to BuildPulse. It returns the
// key that uniquely identifies the uploaded object.
func (s *Submit) Run() (string, error) {
	tarpath, err := s.bundle()
	if err != nil {
		return "", err
	}

	s.logger.Printf("Gzipping tarball (%s)", tarpath)
	zippath, err := toGz(tarpath)
	if err != nil {
		return "", err
	}

	s.logger.Printf("Sending %s to BuildPulse", zippath)
	key, err := s.upload(zippath)
	if err != nil {
		return "", err
	}
	s.logger.Printf("Delivered test results to BuildPulse (%s)", key)

	return key, nil
}

// bundle gathers the artifacts expected by BuildPulse, creates a tarball
// containing those artifacts, and returns the path of the resulting file.
func (s *Submit) bundle() (string, error) {
	// Prepare the metadata file
	//////////////////////////////////////////////////////////////////////////////

	s.logger.Printf("Gathering metadata to describe the build")
	meta, err := metadata.NewMetadata(s.version, s.envs, s.commitResolver, time.Now, s.logger)
	if err != nil {
		return "", err
	}
	yaml, err := meta.MarshalYAML()
	if err != nil {
		return "", err
	}

	yamlfile, err := ioutil.TempFile("", "buildpulse-*.yml")
	if err != nil {
		return "", err
	}
	defer yamlfile.Close()

	s.logger.Printf("Writing metadata to %s", yamlfile.Name())
	_, err = yamlfile.Write(yaml)
	if err != nil {
		return "", err
	}

	// Initialize the tarfile for writing
	//////////////////////////////////////////////////////////////////////////////

	f, err := ioutil.TempFile("", "buildpulse-*.tar")
	if err != nil {
		return "", err
	}
	defer f.Close()

	t := tar.Create(f)
	defer t.Close()

	// Write the XML reports to the tarfile
	//////////////////////////////////////////////////////////////////////////////

	s.logger.Printf("Preparing tarball of test results:")
	for _, p := range s.paths {
		s.logger.Printf("- %s", p)
		err = t.Write(p, p)
		if err != nil {
			return "", err
		}
	}

	// Write the metadata file to the tarfile
	//////////////////////////////////////////////////////////////////////////////

	s.logger.Printf("Adding buildpulse.yml to tarball")
	err = t.Write(yamlfile.Name(), "buildpulse.yml")
	if err != nil {
		return "", err
	}

	// Write the log to the tarfile
	//////////////////////////////////////////////////////////////////////////////

	logfile, err := ioutil.TempFile("", "buildpulse-*.log")
	if err != nil {
		return "", err
	}
	defer logfile.Close()

	s.logger.Printf("Flushing log to %s", logfile.Name())
	_, err = logfile.Write([]byte(s.logger.Text()))
	if err != nil {
		return "", err
	}

	s.logger.Printf("Adding buildpulse.log to tarball")
	err = t.Write(logfile.Name(), "buildpulse.log")
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

// upload transmits the file at the given path to S3
func (s *Submit) upload(path string) (string, error) {
	key := fmt.Sprintf("%d/%d/buildpulse-%s.gz", s.accountID, s.repositoryID, s.idgen())

	err := putS3Object(s.client, s.credentials.AccessKeyID, s.credentials.SecretAccessKey, s.bucket, key, path, s.endpointURL)
	if err != nil {
		return "", err
	}

	return key, nil
}

// toGz gzips the named file (src) and returns the path of the resulting file.
func toGz(src string) (dest string, err error) {
	reader, err := os.Open(src)
	if err != nil {
		return "", err
	}

	zipfile, err := ioutil.TempFile("", "buildpulse-*.gz")
	if err != nil {
		return "", err
	}
	defer zipfile.Close()

	writer := gzip.NewWriter(zipfile)
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	return zipfile.Name(), err
}

// putS3Object puts the named file (src) as an object in the named bucket with the named key.
func putS3Object(client *http.Client, id string, secret string, bucket string, objectKey string, src string, endpointURL string) error {
	provider := &awscreds.StaticProvider{
		Value: awscreds.Value{
			AccessKeyID:     id,
			SecretAccessKey: secret,
		},
	}

	config := aws.NewConfig().
		WithCredentials(awscreds.NewCredentials(provider)).
		WithRegion("us-east-1").
		WithHTTPClient(client).
		WithS3ForcePathStyle(true)

	if len(endpointURL) > 0 {
		config = config.WithEndpoint(endpointURL)
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return err
	}

	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		ACL:    aws.String("bucket-owner-full-control"),
		Body:   file,
	})
	if err != nil {
		return err
	}

	return nil
}

// flagRegex matches args that are flags.
var flagRegex = regexp.MustCompile("^-")

// pathsAndFlagsFromArgs returns a slice containing the subset of args that
// represent paths and a slice containing the subset of args that represent
// flags.
func pathsAndFlagsFromArgs(args []string) ([]string, []string) {
	for i, v := range args {
		isFlag := flagRegex.MatchString(v)

		if isFlag {
			paths := args[0:i]
			flags := args[i:]
			return paths, flags
		}
	}

	return args, []string{}
}

// xmlPathsFromArgs translates each path in args into a list of XML files present
// at that path. It returns the resulting list of XML file paths.
func xmlPathsFromArgs(args []string) ([]string, error) {
	var paths []string

	for _, arg := range args {
		info, err := os.Stat(arg)
		if err == nil && info.IsDir() {
			xmls, err := xmlPathsFromDir(arg)
			if err != nil {
				return nil, err
			}
			paths = append(paths, xmls...)
		} else {
			xmls, err := xmlPathsFromGlob(arg)
			if err != nil {
				return nil, fmt.Errorf("invalid value \"%s\" for path: %v", arg, err)
			}
			paths = append(paths, xmls...)
		}
	}

	return paths, nil
}

// xmlPathsFromDir returns a list of all the XML files in the given directory
// and its subdirectories.
func xmlPathsFromDir(dir string) ([]string, error) {
	var paths []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if isXML(info.Name()) {
			paths = append(paths, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// xmlPathsFromGlob returns a list of all the XML files that match the given
// glob pattern.
func xmlPathsFromGlob(pattern string) ([]string, error) {
	candidates, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, p := range candidates {
		if isXML(p) {
			paths = append(paths, p)
		}
	}

	return paths, nil
}

// isXML returns true if the given filename has an XML extension
// (case-insensitive); false, otherwise.
func isXML(filename string) bool {
	return bytes.EqualFold([]byte(filepath.Ext(filename)), []byte(".xml"))
}
