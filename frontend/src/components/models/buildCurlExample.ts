export type CurlFormat = 'anthropic' | 'openai' | 'gemini'

export interface CurlExample {
  format: CurlFormat
  code: string
}

function providerToFormat(provider: string): CurlFormat {
  if (provider === 'claude') return 'anthropic'
  if (provider === 'gemini') return 'gemini'
  return 'openai'
}

function normalizeBaseRoot(baseUrl: string): string {
  const fallback = baseUrl || (typeof window !== 'undefined' ? window.location.origin : '')
  return fallback.replace(/\/v1\/?$/, '').replace(/\/+$/, '')
}

const PLACEHOLDER_KEY = 'sk-xxxxxx'

function anthropicCurl(modelId: string, baseRoot: string): string {
  return `curl ${baseRoot}/v1/messages \\
  -H "x-api-key: ${PLACEHOLDER_KEY}" \\
  -H "anthropic-version: 2023-06-01" \\
  -H "content-type: application/json" \\
  -d '{
    "model": "${modelId}",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello"}]
  }'`
}

function openaiCurl(modelId: string, baseRoot: string): string {
  return `curl ${baseRoot}/v1/chat/completions \\
  -H "Authorization: Bearer ${PLACEHOLDER_KEY}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${modelId}",
    "messages": [{"role": "user", "content": "Hello"}]
  }'`
}

function geminiCurl(modelId: string, baseRoot: string): string {
  return `curl "${baseRoot}/v1beta/models/${modelId}:generateContent" \\
  -H "x-goog-api-key: ${PLACEHOLDER_KEY}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "contents": [{"parts": [{"text": "Hello"}]}]
  }'`
}

export function buildCurlExample(
  modelId: string,
  provider: string,
  baseUrl: string
): CurlExample {
  const baseRoot = normalizeBaseRoot(baseUrl)
  const format = providerToFormat(provider)

  if (format === 'anthropic') return { format, code: anthropicCurl(modelId, baseRoot) }
  if (format === 'gemini') return { format, code: geminiCurl(modelId, baseRoot) }
  return { format, code: openaiCurl(modelId, baseRoot) }
}
