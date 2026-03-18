import { Hono } from 'hono'

const app = new Hono()

const REPO_URL = 'https://github.com/MSch/capsule'
const INSTALL_SCRIPT_URL = 'https://raw.githubusercontent.com/MSch/capsule/main/scripts/install.sh'
const USER_AGENT = 'capsule-site/1.0'

app.on(['GET', 'HEAD'], '/', (c) => c.redirect(REPO_URL, 302))

app.on(['GET', 'HEAD'], '/install.sh', async (c) => {
  try {
    const upstream = await fetch(INSTALL_SCRIPT_URL, {
      method: c.req.method,
      headers: {
        'user-agent': USER_AGENT,
      },
    })

    if (!upstream.ok) {
      console.error(`install proxy failed with status ${upstream.status}`)
      return c.text('Install script is currently unavailable.\n', 502)
    }

    const headers = new Headers(upstream.headers)
    headers.set('content-type', 'text/x-shellscript; charset=utf-8')
    headers.set('cache-control', 'public, max-age=300')
    headers.set('x-content-type-options', 'nosniff')

    return new Response(c.req.method === 'HEAD' ? null : upstream.body, {
      status: upstream.status,
      headers,
    })
  } catch (error) {
    console.error('install proxy failed', error)
    return c.text('Install script is currently unavailable.\n', 502)
  }
})

app.notFound((c) => {
  if (c.env.ASSETS) {
    return c.env.ASSETS.fetch(c.req.raw)
  }

  return c.text('Not found\n', 404)
})

export default app
