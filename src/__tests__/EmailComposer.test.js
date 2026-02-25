import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import EmailComposer from "@/components/EmailComposer.vue";

// Mock the API
vi.mock("@/api/client-v2.js", () => ({
  post: vi.fn(() => Promise.resolve([{}, null])),
  sendMessage: "/api/send",
}));

describe("EmailComposer", () => {
  const defaultProps = {
    initialSubject: "Test Subject",
    initialBody: "Test body content",
  };

  const mountComponent = (props = defaultProps) => {
    return mount(EmailComposer, {
      props,
      global: {
        stubs: {
          BulmaFieldInput: {
            template:
              '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)">',
            props: ["modelValue", "label"],
            emits: ["update:modelValue"],
          },
          BulmaTextarea: {
            template:
              '<textarea :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)"></textarea>',
            props: ["modelValue", "label", "rows"],
            emits: ["update:modelValue"],
          },
          ErrorSimple: true,
          "font-awesome-icon": true,
        },
      },
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("initializes with provided subject and body", () => {
    const wrapper = mountComponent();
    const input = wrapper.find("input");
    const textarea = wrapper.find("textarea");

    expect(input.element.value).toBe("Test Subject");
    expect(textarea.element.value).toBe("Test body content");
  });

  it("discard button is disabled when content unchanged", () => {
    const wrapper = mountComponent();
    const discardButton = wrapper.findAll("button")[1]; // Second button is discard

    expect(discardButton.attributes("disabled")).toBeDefined();
  });

  it("discard button is enabled when content changes", async () => {
    const wrapper = mountComponent();
    const input = wrapper.find("input");

    await input.setValue("Changed Subject");

    const discardButton = wrapper.findAll("button")[1];
    expect(discardButton.attributes("disabled")).toBeUndefined();
  });

  it("emits hide event when close button clicked", async () => {
    const wrapper = mountComponent();
    const closeButton = wrapper.findAll("button")[2]; // Third button is close

    await closeButton.trigger("click");

    expect(wrapper.emitted("hide")).toBeTruthy();
  });

  it("discard resets to initial values", async () => {
    // Mock window.confirm
    vi.spyOn(window, "confirm").mockReturnValue(true);

    const wrapper = mountComponent();
    const input = wrapper.find("input");
    const textarea = wrapper.find("textarea");

    // Change values
    await input.setValue("Changed Subject");
    await textarea.setValue("Changed body");

    // Click discard
    const discardButton = wrapper.findAll("button")[1];
    await discardButton.trigger("click");

    // Should be back to initial values
    expect(input.element.value).toBe("Test Subject");
    expect(textarea.element.value).toBe("Test body content");
  });

});
