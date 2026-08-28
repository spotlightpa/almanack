import { ref, type Ref } from "vue";
import maybeDate from "@/utils/maybe-date.ts";

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
  updatedAt: Ref<Date | null>; // read-only: populated from server response
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
  toJSON(): Omit<Required<PromotionJSON>, "updated_at">;
}

export function makePromotion(initial: PromotionJSON = {}): Promotion {
  const id = initial.id ?? null;
  const updatedAt = ref<Date | null>(null);
  const name = ref("");
  const description = ref("");
  const width = ref(0);
  const height = ref(0);
  const link = ref("");
  const bannerLabel = ref("");
  const bannerLabelLink = ref("");
  const imageDescription = ref("");
  const imageUrls = ref<string[]>([]);

  function init(src: PromotionJSON = {}): void {
    updatedAt.value = maybeDate(src, "updated_at");
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

  init(initial);

  function toJSON(): Omit<Required<PromotionJSON>, "updated_at"> {
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
