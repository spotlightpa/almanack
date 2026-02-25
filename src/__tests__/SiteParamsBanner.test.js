import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SiteParamsBanner from '@/components/SiteParamsBanner.vue'

describe('SiteParamsBanner', () => {
  const defaultParams = {
    data: {
      'banner-active': false,
      'banner': 'Test banner text',
      'banner-link': '/test-link',
      'banner-bg-color': '#ff6c36',
      'banner-text-color': '#ffffff',
    },
  }

  const mountComponent = (params = defaultParams) => {
    return mount(SiteParamsBanner, {
      props: {
        params,
        fileProps: {},
      },
      global: {
        stubs: {
          BulmaField: {
            template: '<div class="bulma-field"><slot></slot></div>',
            props: ['label', 'help'],
          },
          BulmaTextarea: {
            template: '<textarea class="bulma-textarea" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)"></textarea>',
            props: ['modelValue', 'label', 'help'],
            emits: ['update:modelValue'],
          },
          BulmaFieldInput: {
            template: '<input class="bulma-input" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)">',
            props: ['modelValue', 'label', 'type'],
            emits: ['update:modelValue'],
          },
          BulmaFieldColor: {
            template: '<input type="color" class="bulma-color" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)">',
            props: ['modelValue', 'label', 'help'],
            emits: ['update:modelValue'],
          },
        },
      },
    })
  }

  it('renders banner section title', () => {
    const wrapper = mountComponent()
    expect(wrapper.find('summary').text()).toBe('Banner')
  })

  it('exposes saveData function', () => {
    const wrapper = mountComponent()
    expect(wrapper.vm.saveData).toBeDefined()
    expect(typeof wrapper.vm.saveData).toBe('function')
  })

  it('saveData returns data with correct keys', () => {
    const wrapper = mountComponent()
    const saved = wrapper.vm.saveData()
    
    expect(saved).toHaveProperty('banner-active')
    expect(saved).toHaveProperty('banner')
    expect(saved).toHaveProperty('banner-link')
    expect(saved).toHaveProperty('banner-bg-color')
    expect(saved).toHaveProperty('banner-text-color')
  })

  it('saveData reflects initial values', () => {
    const wrapper = mountComponent()
    const saved = wrapper.vm.saveData()
    
    expect(saved['banner-active']).toBe(false)
    expect(saved['banner-bg-color']).toBe('#ff6c36')
    expect(saved['banner-text-color']).toBe('#ffffff')
  })

  it('checkbox controls banner active state', async () => {
    const wrapper = mountComponent()
    const checkbox = wrapper.find('input[type="checkbox"]')
    
    expect(checkbox.element.checked).toBe(false)
    
    await checkbox.setValue(true)
    
    const saved = wrapper.vm.saveData()
    expect(saved['banner-active']).toBe(true)
  })

  it('banner details are hidden when inactive via v-show', () => {
    const wrapper = mountComponent()
    // v-show renders the element but sets display: none
    const hiddenDiv = wrapper.find('div[style*="display: none"]')
    expect(hiddenDiv.exists()).toBe(true)
  })

  it('banner details are visible when active', async () => {
    const wrapper = mountComponent()
    const checkbox = wrapper.find('input[type="checkbox"]')
    
    await checkbox.setValue(true)
    
    // After activating, the div should no longer have display: none
    const hiddenDiv = wrapper.find('div[style*="display: none"]')
    expect(hiddenDiv.exists()).toBe(false)
  })
})
