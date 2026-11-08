const { execSync } = require('child_process')
const fs = require('fs')
const path = require('path')

// Скрипт не зависит от текущей папки: пути в нём заданы относительно корня проекта.
process.chdir(path.join(__dirname, '..'))

const config = JSON.parse(fs.readFileSync('.claude/ralph.config.json', 'utf8'))

// Сбрасываем счётчик итераций
fs.writeFileSync('.claude/ralph.iterations.json', JSON.stringify({ count: 0, phaseIndex: 0 }))

// Запускаем первую итерацию
const prompt = config.prompt
  .replace('{milestone}', config.phases[0].milestone)
  .replace('{branch}', config.phases[0].branch)
console.log(`🚀 Запускаем Ralph для milestone: ${config.phases[0].milestone}`)

execSync(`claude -p "${prompt}" --max-turns ${config.maxTurns}`, { stdio: 'inherit' })
