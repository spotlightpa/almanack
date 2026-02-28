import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import { ref } from "vue";
import ViewPagesList from "@/components/ViewPagesList.vue";

// Mock vue-router
const mockRoute = ref({
  name: "news-pages",
  meta: {
    contentPath: "content/news/",
    title: "Spotlight PA News Pages",
  },
});

vi.mock("vue-router", () => ({
  useRoute: () => mockRoute.value,
}));

// Mock the API
vi.mock("@/api/client-v2.js", () => ({
  get: vi.fn(),
  listPages: "/api/pages",
}));

// Mock watchAPI to return controlled state
vi.mock("@/api/service-util.js", () => ({
  watchAPI: vi.fn(() => ({
    apiState: ref({ isLoading: false, error: null }),
    fetch: vi.fn(),
    computedList: () => ref([]),
    computedProp: () => ref(null),
  })),
}));

describe("ViewPagesList", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uses route.meta.title for the page title", () => {
    mockRoute.value = {
      name: "news-pages",
      meta: {
        contentPath: "content/news/",
        title: "Spotlight PA News Pages",
      },
    };

    const wrapper = mount(ViewPagesList, {
      props: { page: "" },
      global: {
        stubs: {
          MetaHead: {
            template: "<div><slot></slot></div>",
          },
          PageList: true,
        },
      },
    });

    expect(wrapper.text()).toContain("Spotlight PA News Pages");
  });

  it("passes title from route.meta to PageList", () => {
    mockRoute.value = {
      name: "statecollege-pages",
      meta: {
        contentPath: "content/statecollege/",
        title: "State College Pages",
      },
    };

    const wrapper = mount(ViewPagesList, {
      props: { page: "" },
      global: {
        stubs: {
          MetaHead: true,
          PageList: {
            template: '<div class="page-list">{{ title }}</div>',
            props: ["title", "page", "nextPage", "apiState", "reload", "pages"],
          },
        },
      },
    });

    expect(wrapper.find(".page-list").text()).toBe("State College Pages");
  });

  it("works with different route configurations", () => {
    // Test that the component adapts to different routes
    const routes = [
      {
        name: "news-pages",
        title: "Spotlight PA News Pages",
        path: "content/news/",
      },
      {
        name: "statecollege-pages",
        title: "State College Pages",
        path: "content/statecollege/",
      },
      { name: "berks-pages", title: "Berks Pages", path: "content/berks/" },
      {
        name: "sponsored-pages",
        title: "Sponsored Content",
        path: "content/sponsored/",
      },
    ];

    for (const routeConfig of routes) {
      mockRoute.value = {
        name: routeConfig.name,
        meta: {
          contentPath: routeConfig.path,
          title: routeConfig.title,
        },
      };

      const wrapper = mount(ViewPagesList, {
        props: { page: "" },
        global: {
          stubs: {
            MetaHead: {
              template: "<div><slot></slot></div>",
            },
            PageList: true,
          },
        },
      });

      expect(wrapper.text()).toContain(routeConfig.title);
    }
  });
});
