import { ref, type Ref } from "vue";

export interface PromotionJSON {
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
  updated_at?: string;
}

export interface Promotion {
  id: number | null;
  updatedAt: Ref<Date | null>;
  name: Ref<string>;
  description: Ref<string>;
  width: Ref<number>;
  height: Ref<number>;
  link: Ref<string>;
  bannerLabel: Ref<string>;
  bannerLabelLink: Ref<string>;
  imageDescription: Ref<string>;
  imageUrls: Ref<string[]>;
  init(src?: PromotionJSON): void;
  toJSON(): Required<PromotionJSON>;
}

export function makePromotion(initial: PromotionJSON = {}): Promotion {
  const id = initial.id ?? null;
  const updatedAt = ref<Date | null>(
    initial.updated_at ? new Date(initial.updated_at) : null
  );
  const name = ref(initial.name ?? "");
  const description = ref(initial.description ?? "");
  const width = ref(initial.width ?? 0);
  const height = ref(initial.height ?? 0);
  const link = ref(initial.link ?? "");
  const bannerLabel = ref(initial.banner_label ?? "");
  const bannerLabelLink = ref(initial.banner_label_link ?? "");
  const imageDescription = ref(initial.image_description ?? "");
  const imageUrls = ref([...(initial.image_urls ?? [])]);

  function init(src: PromotionJSON = {}): void {
    updatedAt.value = src.updated_at ? new Date(src.updated_at) : null;
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

  function toJSON(): Required<PromotionJSON> {
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
      updated_at: updatedAt.value?.toISOString() ?? "",
    };
  }

  return {
    id,
    updatedAt,
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
