import { ref } from "vue";

export function makePromotion(initial = {}) {
  const name = ref(initial.name ?? "");
  const description = ref(initial.description ?? "");
  const width = ref(initial.width ?? 0);
  const height = ref(initial.height ?? 0);
  const link = ref(initial.link ?? "");
  const bannerLabel = ref(initial.banner_label ?? "");
  const bannerLabelLink = ref(initial.banner_label_link ?? "");
  const imageDescription = ref(initial.image_description ?? "");
  const imageUrls = ref([...(initial.image_urls ?? [])]);

  function init(src = {}) {
    name.value = src.name ?? "";
    description.value = src.description ?? "";
    width.value = src.width ?? 0;
    height.value = src.height ?? 0;
    link.value = src.link ?? "";
    bannerLabel.value = src.banner_label ?? "";
    bannerLabelLink.value = src.banner_label_link ?? "";
    imageDescription.value = src.image_description ?? "";
    imageUrls.value = [...(src.image_urls ?? [])];
  }

  function toJSON(id) {
    return {
      id: id ?? null,
      name: name.value,
      description: description.value,
      width: width.value,
      height: height.value,
      link: link.value,
      banner_label: bannerLabel.value,
      banner_label_link: bannerLabelLink.value,
      image_description: imageDescription.value,
      image_urls: imageUrls.value,
    };
  }

  return {
    name,
    description,
    width,
    height,
    link,
    bannerLabel,
    bannerLabelLink,
    imageDescription,
    imageUrls,
    init,
    toJSON,
  };
}
