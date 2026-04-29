// Tailwind class strings shared across auth views.
// Inline-tailwind is verbose; centralizing the input/label/error/cta tokens here
// keeps each form template readable while preserving the same visual language as
// the public homepage chrome (border-only, neutral grays, blue/violet hover glow).

const INPUT_BASE =
  'block w-full rounded-md border bg-white px-3 py-2.5 text-sm text-gray-900 ' +
  'placeholder-gray-400 transition-colors focus:outline-none focus:ring-1 ' +
  'disabled:cursor-not-allowed disabled:bg-gray-50 disabled:opacity-70 ' +
  'dark:bg-gray-950 dark:text-white dark:placeholder-gray-600 ' +
  'dark:disabled:bg-gray-900'

const INPUT_NORMAL =
  'border-gray-200 focus:border-blue-500 focus:ring-blue-500 ' +
  'dark:border-gray-800 dark:focus:border-blue-500'

const INPUT_ERROR =
  'border-red-500 focus:border-red-500 focus:ring-red-500 ' +
  'dark:border-red-500/70 dark:focus:border-red-500'

const INPUT_VALID =
  'border-emerald-500 focus:border-emerald-500 focus:ring-emerald-500 ' +
  'dark:border-emerald-500/70 dark:focus:border-emerald-500'

export function inputClass(error = false, valid = false): string {
  if (error) return `${INPUT_BASE} ${INPUT_ERROR}`
  if (valid) return `${INPUT_BASE} ${INPUT_VALID}`
  return `${INPUT_BASE} ${INPUT_NORMAL}`
}
