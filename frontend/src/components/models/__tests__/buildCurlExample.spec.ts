import { describe, expect, it } from 'vitest'
import { buildCurlExample } from '../buildCurlExample'

describe('buildCurlExample', () => {
  it('builds Anthropic format for claude provider', () => {
    const r = buildCurlExample('claude-opus-4-6', 'claude', 'https://api.example.com')
    expect(r.format).toBe('anthropic')
    expect(r.code).toContain('https://api.example.com/v1/messages')
    expect(r.code).toContain('"model": "claude-opus-4-6"')
    expect(r.code).toContain('x-api-key: sk-xxxxxx')
    expect(r.code).toContain('anthropic-version: 2023-06-01')
    expect(r.code).toContain('"max_tokens": 1024')
  })

  it('builds OpenAI format for openai provider', () => {
    const r = buildCurlExample('gpt-5', 'openai', 'https://api.example.com')
    expect(r.format).toBe('openai')
    expect(r.code).toContain('https://api.example.com/v1/chat/completions')
    expect(r.code).toContain('"model": "gpt-5"')
    expect(r.code).toContain('Authorization: Bearer sk-xxxxxx')
  })

  it('builds Gemini format for gemini provider', () => {
    const r = buildCurlExample('gemini-2.0-flash', 'gemini', 'https://api.example.com')
    expect(r.format).toBe('gemini')
    expect(r.code).toContain('https://api.example.com/v1beta/models/gemini-2.0-flash:generateContent')
    expect(r.code).toContain('x-goog-api-key: sk-xxxxxx')
    expect(r.code).toContain('"contents": [{"parts": [{"text": "Hello"}]}]')
  })

  it('strips trailing /v1 from baseUrl', () => {
    const r = buildCurlExample('claude-opus', 'claude', 'https://api.example.com/v1')
    expect(r.code).toContain('https://api.example.com/v1/messages')
    expect(r.code).not.toContain('/v1/v1')
  })

  it('strips trailing /v1/ from baseUrl', () => {
    const r = buildCurlExample('gpt-5', 'openai', 'https://api.example.com/v1/')
    expect(r.code).toContain('https://api.example.com/v1/chat/completions')
    expect(r.code).not.toContain('/v1/v1')
  })

  it('strips trailing slash from baseUrl', () => {
    const r = buildCurlExample('gpt-5', 'openai', 'https://api.example.com/')
    expect(r.code).toContain('https://api.example.com/v1/chat/completions')
  })

  it('falls back to OpenAI format for unknown provider', () => {
    const r = buildCurlExample('mystery-model', 'unknown', 'https://api.example.com')
    expect(r.format).toBe('openai')
    expect(r.code).toContain('/v1/chat/completions')
  })
})
