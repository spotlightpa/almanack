import { ref, type Ref } from "vue";

export interface PromotionData {
  id?: number | null;
  name?: string;
  description?: string;
  width?: number;
  height?: number;
  link?: string;
  banner_label?: string;
  banner_label_link?: string;
  image_description?: string;
  image_urls?: string[];
}

export interface Promotion {
  name: Ref<string>;
  description: Ref<string>;
  width: Ref<number>;
  height: Ref<number>;
  link: Ref<string>;
  bannerLabel: Ref<string>;
  bannerLabelLink: Ref<string>;
  imageDescription: Ref<string>;
  imageUrls: Ref<string[]>;
  init(src?: PromotionData): void;
  toJSON(id: number | null): Required<PromotionData>;
}

export function makePromotion(initial: PromotionData = {}): Promotion {
  const name = ref(initial.name ?? "");
  const description = ref(initial.description ?? "");
  const width = ref(initial.width ?? 0);
  const height = ref(initial.height ?? 0);
  const link = ref(initial.link ?? "");
  const bannerLabel = ref(initial.banner_label ?? "");
  const bannerLabelLink = ref(initial.banner_label_link ?? "");
  const imageDescription = ref(initial.image_description ?? "");
  const imageUrls = ref([...(initial.image_urls ?? [])]);

  function init(src: PromotionData = {}): void {
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

  function toJSON(id: number | null): Required<PromotionData> {
    return {
      id,
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
