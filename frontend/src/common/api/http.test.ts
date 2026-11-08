import { ApiError, errorMessage, http, NETWORK_ERROR } from '@/common/api/http'

function respond(status: number, body?: unknown) {
  const text = body === undefined ? '' : typeof body === 'string' ? body : JSON.stringify(body)
  // Ответ 204 по спецификации не может иметь тела, даже пустого.
  return vi.fn<typeof fetch>().mockResolvedValue(new Response(status === 204 ? null : text, { status }))
}

describe('http', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('отправляет JSON и заголовок сессии', async () => {
    const fetchMock = respond(200, { ok: true })
    vi.stubGlobal('fetch', fetchMock)

    await http('/api/x', { method: 'POST', body: { a: 1 }, sessionUuid: 'sess' })

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    const headers = init.headers as Headers
    expect(init.method).toBe('POST')
    expect(init.body).toBe('{"a":1}')
    expect(headers.get('Content-Type')).toBe('application/json')
    expect(headers.get('X-Session-Uuid')).toBe('sess')
  })

  it('не добавляет Content-Type без тела и сессии', async () => {
    const fetchMock = respond(200, {})
    vi.stubGlobal('fetch', fetchMock)

    await http('/api/x')

    const headers = (fetchMock.mock.calls[0] as [string, RequestInit])[1].headers as Headers
    expect(headers.has('Content-Type')).toBe(false)
    expect(headers.has('X-Session-Uuid')).toBe(false)
  })

  it('возвращает undefined на 204', async () => {
    vi.stubGlobal('fetch', respond(204))

    await expect(http('/api/x', { method: 'POST' })).resolves.toBeUndefined()
  })

  it.each([
    ['{message}', { message: 'из message' }, 'из message'],
    ['{error_message}', { error_message: 'из error_message' }, 'из error_message'],
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
