import { ApiError, errorMessage, http, NETWORK_ERROR } from '@/common/api/http'

function respond(status: number, body?: unknown) {
  const text = body === undefined ? '' : typeof body === 'string' ? body : JSON.stringify(body)
  // Ответ 204 по спецификации не может иметь тела, даже пустого.
  return vi.fn<typeof fetch>().mockResolvedValue(new Response(status === 204 ? null : text, { status }))
}

/** ky вызывает fetch с готовым Request. */
function sentRequest(fetchMock: { mock: { calls: unknown[][] } }): Request {
  const [request] = fetchMock.mock.calls[0] ?? []
  if (!(request instanceof Request)) throw new Error('fetch не вызван с Request')
  return request
}

describe('http', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('отправляет JSON и заголовок сессии', async () => {
    // Тело читаем внутри мока: после вызова fetch Request уже использован.
    let sentBody = ''
    const fetchMock = vi.fn<typeof fetch>(async (input) => {
      sentBody = await (input as Request).clone().text()
      return new Response('{}', { status: 200 })
    })
    vi.stubGlobal('fetch', fetchMock)

    await http('/api/x', { method: 'POST', body: { a: 1 }, sessionUuid: 'sess' })

    const request = sentRequest(fetchMock)
    expect(request.method).toBe('POST')
    expect(sentBody).toBe('{"a":1}')
    expect(request.headers.get('Content-Type')).toBe('application/json')
    expect(request.headers.get('Session-Uuid')).toBe('sess')
  })

  it('не добавляет Content-Type без тела и сессии', async () => {
    const fetchMock = respond(200, {})
    vi.stubGlobal('fetch', fetchMock)

    await http('/api/x')

    const { headers } = sentRequest(fetchMock)
    expect(headers.has('Content-Type')).toBe(false)
    expect(headers.has('Session-Uuid')).toBe(false)
  })

  it('возвращает undefined на 204', async () => {
    vi.stubGlobal('fetch', respond(204))

    await expect(http('/api/x', { method: 'POST' })).resolves.toBeUndefined()
  })

  it.each([
    ['{message}', { message: 'из message' }, 'из message'],
    ['{error_message}', { error_message: 'из error_message' }, 'из error_message'],
    ['{code, message, details} от Envoy', { code: 16, message: 'invalid session', details: [] }, 'invalid session'],
    ['{error: {message}}', { error: { code: 'X', message: 'вложенное' } }, 'вложенное'],
  ])('достаёт сообщение об ошибке из формы %s', async (_name, body, expected) => {
    vi.stubGlobal('fetch', respond(400, body))

    await expect(http('/api/x')).rejects.toMatchObject({ status: 400, message: expected })
  })

  it('подставляет сообщение по умолчанию, если тело не разобрать', async () => {
    vi.stubGlobal('fetch', respond(502, '<html>bad gateway</html>'))

    await expect(http('/api/x')).rejects.toMatchObject({ status: 502, message: 'Ошибка 502' })
  })

  it('превращает сетевой сбой в ApiError со статусом 0', async () => {
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockRejectedValue(new TypeError('Failed to fetch')))

    await expect(http('/api/x')).rejects.toMatchObject({ status: NETWORK_ERROR })
  })

  it('пробрасывает AbortError как есть, чтобы отмену не показывали как ошибку', async () => {
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockRejectedValue(new DOMException('aborted', 'AbortError')))

    await expect(http('/api/x')).rejects.toMatchObject({ name: 'AbortError' })
  })
})

describe('errorMessage', () => {
  it('берёт сообщение ApiError и подставляет запасное для остальных', () => {
    expect(errorMessage(new ApiError(400, 'плохой запрос'), 'запасное')).toBe('плохой запрос')
    expect(errorMessage(new Error('внутреннее'), 'запасное')).toBe('запасное')
    expect(errorMessage(undefined, 'запасное')).toBe('запасное')
  })
})
