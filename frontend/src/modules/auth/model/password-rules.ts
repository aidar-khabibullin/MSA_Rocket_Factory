// Правила iam (iam/internal/model/user.go): 8-128 символов, заглавная, строчная, цифра, спецсимвол.
export const PASSWORD_RULES = [
  { id: 'length', labelKey: 'passwordRules.length', test: (value: string) => value.length >= 8 },
  { id: 'upper', labelKey: 'passwordRules.upper', test: (value: string) => /\p{Lu}/u.test(value) },
  { id: 'lower', labelKey: 'passwordRules.lower', test: (value: string) => /\p{Ll}/u.test(value) },
  { id: 'digit', labelKey: 'passwordRules.digit', test: (value: string) => /\p{N}/u.test(value) },
  { id: 'special', labelKey: 'passwordRules.special', test: (value: string) => /[\p{P}\p{S}]/u.test(value) },
] as const

export const isPasswordValid = (value: string): boolean => PASSWORD_RULES.every((rule) => rule.test(value))

// Ошибки возвращаются ключами словаря auth: перевод выполняет UI, а модель остаётся независимой от языка.
export const passwordError = (value: string): 'validation.passwordInvalid' | null =>
  isPasswordValid(value) ? null : 'validation.passwordInvalid'

export const loginError = (value: string): 'validation.loginMin' | null =>
  value.trim().length < 3 ? 'validation.loginMin' : null
