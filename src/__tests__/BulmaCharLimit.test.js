import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BulmaCharLimit from '@/components/BulmaCharLimit.vue'

describe('BulmaCharLimit', () => {
  it('does not render when under warn threshold', () => {
    const wrapper = mount(BulmaCharLimit, {
      props: { value: 'hello', max: 20, warn: 15 },
    })
    expect(wrapper.find('div').exists()).toBe(false)
  })

  it('renders when over warn threshold', () => {
    const wrapper = mount(BulmaCharLimit, {
      props: { value: 'a'.repeat(16), max: 20, warn: 15 },
    })
    expect(wrapper.find('div').exists()).toBe(true)
    expect(wrapper.text()).toContain('16')
  })

  it('shows warning progress when near limit', () => {
    const wrapper = mount(BulmaCharLimit, {
      props: { value: 'a'.repeat(16), max: 20, warn: 15 },
    })
    expect(wrapper.find('progress').classes()).toContain('is-warning')
  })

  it('shows danger progress when over limit', () => {
    const wrapper = mount(BulmaCharLimit, {
      props: { value: 'a'.repeat(21), max: 20, warn: 15 },
    })
    expect(wrapper.find('progress').classes()).toContain('is-danger')
  })
})
