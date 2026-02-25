import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import HomepageEditor from '@/components/HomepageEditor.vue'

describe('HomepageEditor', () => {
  const createEditorsPicks = () => ({
    featuredStories: [],
    subfeatures: [],
    topSlots: [],
    edCallout: [],
    edInvestigations: [],
    edImpact: [],
  })

  const globalStubs = {
    PageFinder: {
      name: 'PageFinder',
      template: '<div class="page-finder-stub"></div>',
      emits: ['select-page'],
    },
    HomepageEditorDraggable: true,
    BulmaField: {
      template: '<div class="bulma-field-stub"><slot></slot></div>',
      props: ['label'],
    },
  }

  it('renders PageFinder component', () => {
    const wrapper = mount(HomepageEditor, {
      props: {
        editorsPicks: createEditorsPicks(),
        showCallout: false,
        showInvestigation: false,
        showImpact: false,
      },
      global: { stubs: globalStubs },
    })
    expect(wrapper.find('.page-finder-stub').exists()).toBe(true)
  })

  it('conditional fields appear based on props', () => {
    const wrapperWithAll = mount(HomepageEditor, {
      props: {
        editorsPicks: createEditorsPicks(),
        showCallout: true,
        showInvestigation: true,
        showImpact: true,
      },
      global: { stubs: globalStubs },
    })
    const wrapperWithNone = mount(HomepageEditor, {
      props: {
        editorsPicks: createEditorsPicks(),
        showCallout: false,
        showInvestigation: false,
        showImpact: false,
      },
      global: { stubs: globalStubs },
    })
    
    // With all flags true, there should be more BulmaField stubs
    const fieldsWithAll = wrapperWithAll.findAll('.bulma-field-stub')
    const fieldsWithNone = wrapperWithNone.findAll('.bulma-field-stub')
    expect(fieldsWithAll.length).toBeGreaterThan(fieldsWithNone.length)
  })

  it('push function adds article to featuredStories via event', async () => {
    const editorsPicks = createEditorsPicks()
    const wrapper = mount(HomepageEditor, {
      props: {
        editorsPicks,
        showCallout: false,
        showInvestigation: false,
        showImpact: false,
      },
      global: { stubs: globalStubs },
    })

    // Find the PageFinder stub and emit select-page
    const pageFinder = wrapper.findComponent({ name: 'PageFinder' })
    await pageFinder.vm.$emit('select-page', { filePath: '/content/news/test-article.md' })

    expect(editorsPicks.featuredStories).toContain('/content/news/test-article.md')
  })

  it('multiple articles can be added', async () => {
    const editorsPicks = createEditorsPicks()
    const wrapper = mount(HomepageEditor, {
      props: {
        editorsPicks,
        showCallout: false,
        showInvestigation: false,
        showImpact: false,
      },
      global: { stubs: globalStubs },
    })

    const pageFinder = wrapper.findComponent({ name: 'PageFinder' })
    await pageFinder.vm.$emit('select-page', { filePath: '/content/news/article1.md' })
    await pageFinder.vm.$emit('select-page', { filePath: '/content/news/article2.md' })

    expect(editorsPicks.featuredStories).toHaveLength(2)
    expect(editorsPicks.featuredStories).toContain('/content/news/article1.md')
    expect(editorsPicks.featuredStories).toContain('/content/news/article2.md')
  })
})
