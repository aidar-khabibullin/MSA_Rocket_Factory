import { isPasswordValid, loginError, PASSWORD_RULES, passwordError } from './password-rules'

describe('passwordRules', () => {
  it.each([
    ['Secret123!', true],
    ['Пароль123!', true], // кириллица тоже буква: правила iam основаны на unicode
    ['short1!A', true],
    ['Short1!', false], // меньше 8
    ['secret123!', false], // нет заглавной
    ['SECRET123!', false], // нет строчной
    ['Secret!!!!', false], // нет цифры
    ['Secret1234', false], // нет спецсимвола
    ['', false],
  ])('%s -> %s', (password, expected) => {
    expect(isPasswordValid(password)).toBe(expected)
    expect(passwordError(password)).toBe(expected ? null : 'validation.passwordInvalid')
  })

  it('каждое правило проверяется независимо', () => {
    const results = Object.fromEntries(PASSWORD_RULES.map((rule) => [rule.id, rule.test('Abcd1')]))

    expect(results).toEqual({ length: false, upper: true, lower: true, digit: true, special: false })
  })

  it('логин: минимум 3 символа без учёта пробелов по краям', () => {
    expect(loginError('ab')).not.toBeNull()
    expect(loginError('  ab  ')).not.toBeNull()
    expect(loginError('abc')).toBeNull()
  })
})
