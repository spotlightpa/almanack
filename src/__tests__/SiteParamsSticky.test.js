import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SiteParamsSticky from '@/components/SiteParamsSticky.vue'

describe('SiteParamsSticky', () => {
  const defaultParams = {
    data: {
      'sticky-active': true,
      'sticky-image-description': 'Test description',
      'sticky-images': ['image1.jpg', 'image2.jpg'],
      'sticky-link': '/sticky-link',
    },
  }

  const mountComponent = (params = defaultParams) => {
    return mount(SiteParamsSticky, {
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
          BulmaFieldInput: {
            template: '<input class="bulma-input" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)">',
            props: ['modelValue', 'label', 'type'],
            emits: ['update:modelValue'],
          },
          BulmaTextarea: true,
          ImageSetInputs: true,
          SiteParamsFiles: true,
        },
      },
    })
  }

  it('exposes saveData function', () => {
    const wrapper = mountComponent()
    expect(wrapper.vm.saveData).toBeDefined()
  })

  it('saveData returns all sticky properties', () => {
    const wrapper = mountComponent()
    const saved = wrapper.vm.saveData()
    
    expect(saved['sticky-active']).toBe(true)
    expect(saved['sticky-image-description']).toBe('Test description')
    expect(saved['sticky-images']).toEqual(['image1.jpg', 'image2.jpg'])
    expect(saved['sticky-link']).toBe('/sticky-link')
  })

  it('toggling checkbox updates saved state', async () => {
    const wrapper = mountComponent()
    const checkbox = wrapper.find('input[type="checkbox"]')
    
    expect(wrapper.vm.saveData()['sticky-active']).toBe(true)
    
    await checkbox.setValue(false)
    
    expect(wrapper.vm.saveData()['sticky-active']).toBe(false)
  })

  it('renders with inactive state correctly', () => {
    const inactiveParams = {
      data: {
        'sticky-active': false,
        'sticky-image-description': '',
        'sticky-images': [],
        'sticky-link': '',
      },
    }
    const wrapper = mountComponent(inactiveParams)
    const saved = wrapper.vm.saveData()
    
    expect(saved['sticky-active']).toBe(false)
    expect(saved['sticky-images']).toEqual([])
  })
})
