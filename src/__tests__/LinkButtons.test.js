import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import LinkButtons from "@/components/LinkButtons.vue";

describe("LinkButtons", () => {
  it("renders label", () => {
    const wrapper = mount(LinkButtons, {
      props: { label: "Actions" },
    });
    expect(wrapper.find(".label").text()).toBe("Actions");
  });

  it("renders slot content in buttons nav", () => {
    const wrapper = mount(LinkButtons, {
      props: { label: "Test" },
      slots: {
        default: "<button>Button 1</button><button>Button 2</button>",
      },
    });
    const buttons = wrapper.findAll("button");
    expect(buttons).toHaveLength(2);
  });
});
