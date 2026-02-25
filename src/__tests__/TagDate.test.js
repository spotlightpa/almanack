import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import TagDate from '@/components/TagDate.vue'

describe('TagDate', () => {
  it('renders formatted date', () => {
    const testDate = new Date('2024-06-15T12:00:00')
    const wrapper = mount(TagDate, {
      props: { date: testDate },
    })
    // formatDate returns something like "Jun 15, 2024"
    expect(wrapper.text()).toContain('Jun')
    expect(wrapper.text()).toContain('15')
    expect(wrapper.text()).toContain('2024')
  })

  it('has tag class', () => {
    const wrapper = mount(TagDate, {
      props: { date: new Date() },
    })
    expect(wrapper.classes()).toContain('tag')
  })
})
