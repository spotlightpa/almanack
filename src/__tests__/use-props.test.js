import { describe, it, expect } from 'vitest'
import { isRef } from 'vue'
import useProps from '@/utils/use-props.js'

describe('useProps', () => {
  it('creates refs from source object properties', () => {
    const src = { 'my-prop': 'hello' }
    const [data] = useProps(src, {
      myProp: ['my-prop'],
    })
    
    expect(isRef(data.myProp)).toBe(true)
    expect(data.myProp.value).toBe('hello')
  })

  it('applies deserialize transform on init', () => {
    const src = { count: '42' }
    const [data] = useProps(src, {
      count: ['count', (v) => parseInt(v, 10)],
    })
    
    expect(data.count.value).toBe(42)
  })

  it('saveData returns serialized values', () => {
    const src = { name: 'test' }
    const [data, saveData] = useProps(src, {
      name: ['name'],
    })
    
    data.name.value = 'changed'
    const result = saveData()
    
    expect(result).toEqual({ name: 'changed' })
  })

  it('saveData applies serialize transform', () => {
    const src = { count: 42 }
    const [data, saveData] = useProps(src, {
      count: ['count', undefined, (v) => String(v)],
    })
    
    data.count.value = 100
    const result = saveData()
    
    expect(result).toEqual({ count: '100' })
  })

  it('handles multiple props with different transforms', () => {
    const src = {
      'is-active': true,
      'link-url': '/path/to/page',
      'text-content': '<b>bold</b>',
    }
    
    const toAbs = (v) => v ? `https://example.com${v}` : ''
    const toRel = (v) => v?.replace('https://example.com', '') ?? ''
    const sanitize = (v) => v?.replace(/<[^>]*>/g, '') ?? ''
    
    const [data, saveData] = useProps(src, {
      isActive: ['is-active'],
      link: ['link-url', toAbs, toRel],
      text: ['text-content', undefined, sanitize],
    })
    
    expect(data.isActive.value).toBe(true)
    expect(data.link.value).toBe('https://example.com/path/to/page')
    expect(data.text.value).toBe('<b>bold</b>')
    
    const result = saveData()
    expect(result['is-active']).toBe(true)
    expect(result['link-url']).toBe('/path/to/page')
    expect(result['text-content']).toBe('bold')
  })

  it('handles undefined source values', () => {
    const src = {}
    const [data, saveData] = useProps(src, {
      missing: ['missing-prop'],
    })
    
    expect(data.missing.value).toBeUndefined()
    
    const result = saveData()
    expect(result['missing-prop']).toBeUndefined()
  })
})
