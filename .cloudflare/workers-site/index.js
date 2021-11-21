import { getAssetFromKV, mapRequestToAsset } from '@cloudflare/kv-asset-handler'

/**
 * The DEBUG flag will do two things that help during development:
 * 1. we will skip caching on the edge, which makes it easier to
 *    debug.
 * 2. we will return an error message on exception in your Response rather
 *    than the default 404.html page.
 */
const DEBUG = false

addEventListener('fetch', event => {
  event.respondWith(handleEvent(event))
})

async function handleEvent(event) {
  let options = {}

  try {
    if (DEBUG) {
      // customize caching
      options.cacheControl = {
        bypassCache: true,
      }
    }

    const pathname = new URL(event.request.url).pathname

    // if the root URL is requested, redirect to the latest release on GitHub
    if (pathname === '/') {
      return Response.redirect('https://github.com/buildpulse/test-reporter/releases/latest', 302)
    }

    const page = await getAssetFromKV(event, options)

    // allow headers to be altered
    const response = new Response(page.body, page)
    response.headers.set('X-XSS-Protection', '1; mode=block')
    response.headers.set('X-Content-Type-Options', 'nosniff')
    response.headers.set('X-Frame-Options', 'DENY')
    response.headers.set('Referrer-Policy', 'unsafe-url')
    response.headers.set('Feature-Policy', 'none')

    const pathnameWithoutLeadingSlash = pathname.substring(1)
    if (isTestReporterBinary(pathnameWithoutLeadingSlash)) {
      response.headers.set('Content-Type', 'application/octet-stream')
      response.headers.set('Content-Disposition', `attachment; filename=${pathnameWithoutLeadingSlash}`)
    }

    return response

  } catch (e) {
    // if an error is thrown try to serve the asset at 404.html
    if (!DEBUG) {
      try {
        let notFoundResponse = await getAssetFromKV(event, {
          mapRequestToAsset: req => new Request(`${new URL(req.url).origin}/404.html`, req),
        })

        return new Response(notFoundResponse.body, { ...notFoundResponse, status: 404 })
      } catch (e) {}
    }

    return new Response(e.message || e.toString(), { status: 500 })
  }
}

function isTestReporterBinary(filename) {
  const kebabCaseFileRegex = /test-reporter-[a-z0-9]+-[a-z0-9]+$/
  const snakeCaseFileRegex = /test_reporter_[a-z0-9]+_[a-z0-9]+$/

  return filename.match(kebabCaseFileRegex) || filename.match(snakeCaseFileRegex)
}
