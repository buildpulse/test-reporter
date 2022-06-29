FROM alpine:3.16

RUN apk add curl

RUN curl -fsSL --retry 3 --retry-connrefused https://get.buildpulse.io/test-reporter-linux-amd64 > ./buildpulse-test-reporter

RUN chmod +x ./buildpulse-test-reporter
