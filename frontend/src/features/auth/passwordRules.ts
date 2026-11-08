// Правила iam (iam/internal/model/user.go): 8-128 символов, заглавная, строчная, цифра, спецсимвол.
export const PASSWORD_RULES = [
  { id: 'length', label: 'Не менее 8 символов', test: (value: string) => value.length >= 8 },
  { id: 'upper', label: 'Заглавная буква', test: (value: string) => /\p{Lu}/u.test(value) },
  { id: 'lower', label: 'Строчная буква', test: (value: string) => /\p{Ll}/u.test(value) },
  { id: 'digit', label: 'Цифра', test: (value: string) => /\p{N}/u.test(value) },
  { id: 'special', label: 'Спецсимвол (! ? #)', test: (value: string) => /[\p{P}\p{S}]/u.test(value) },
] as const

export const isPasswordValid = (value: string): boolean => PASSWORD_RULES.every((rule) => rule.test(value))

export const passwordError = (value: string): string | null =>
  isPasswordValid(value) ? null : 'Пароль не соответствует требованиям'

export const loginError = (value: string): string | null =>
  value.trim().length < 3 ? 'Минимум 3 символа' : null
