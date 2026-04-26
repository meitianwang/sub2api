export interface ModelEntry {
  model_id: string
  display_name: string
  provider: string
  group_id: number
  group_name: string
  input_price: number
  output_price: number
}

export function providerLabel(provider: string): string {
  if (provider === 'claude') return 'Anthropic'
  if (provider === 'openai') return 'OpenAI'
  if (provider === 'gemini') return 'Google'
  return provider
}

export function fmtPrice(p: number): string {
  if (p < 0.01) return p.toFixed(4)
  if (p < 1) return p.toFixed(3)
  return p.toFixed(2)
}
