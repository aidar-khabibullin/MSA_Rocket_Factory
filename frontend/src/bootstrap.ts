import { migrateLegacyStorage } from '@/shared/lib/storage'

// Единоразовая инициализация (правило advanced-init-once). Сторы Zustand читают localStorage в момент
// создания, то есть при импорте, поэтому этот модуль подключается в main.tsx ПЕРВЫМ: ES-модули
// вычисляются в порядке импортов, и перенос данных старой версии успевает выполниться до создания сторов.
migrateLegacyStorage()
