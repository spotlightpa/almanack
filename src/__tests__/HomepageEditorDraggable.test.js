import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import HomepageEditorDraggable from '@/components/HomepageEditorDraggable.vue'

describe('HomepageEditorDraggable', () => {
  const mountComponent = (modelValue = []) => {
    return mount(HomepageEditorDraggable, {
      props: { modelValue },
      global: {
        stubs: {
          draggable: {
            template: `
              <div class="draggable-stub">
                <slot name="header"></slot>
                <template v-for="(element, index) in list" :key="index">
                  <slot name="item" :element="element" :index="index"></slot>
                </template>
              </div>
            `,
            props: ['list', 'itemKey', 'group', 'ghostClass', 'chosenClass', 'componentData'],
          },
          HomepageEditorItem: {
            template: '<div class="item-stub" @click="$emit(\'remove\')">{{ filePath }}</div>',
            props: ['filePath'],
            emits: ['remove'],
          },
        },
      },
    })
  }

  it('shows placeholder when empty', () => {
    const wrapper = mountComponent([])
    expect(wrapper.text()).toContain('Drag articles here')
  })

  it('does not show placeholder when has items', () => {
    const wrapper = mountComponent(['/content/news/article1.md'])
    expect(wrapper.text()).not.toContain('Drag articles here')
  })

  it('renders items from modelValue', () => {
    const wrapper = mountComponent([
      '/content/news/article1.md',
      '/content/news/article2.md',
    ])
    const items = wrapper.findAll('.item-stub')
    expect(items).toHaveLength(2)
  })

  it('emits update:modelValue when item is removed', async () => {
    const wrapper = mountComponent([
      '/content/news/article1.md',
      '/content/news/article2.md',
      '/content/news/article3.md',
    ])
    
    // Click the second item to trigger remove
    const items = wrapper.findAll('.item-stub')
    await items[1].trigger('click')
    
    const emitted = wrapper.emitted('update:modelValue')
    expect(emitted).toBeTruthy()
    expect(emitted[0][0]).toEqual([
      '/content/news/article1.md',
      '/content/news/article3.md',
    ])
  })
})
