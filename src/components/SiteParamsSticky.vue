<script>
import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    let [data, saveData] = useProps(props.params.data, {
      stickyActive: ["sticky-active"],
      stickyImageDescription: ["sticky-image-description"],
      stickyImages: ["sticky-images"],
      stickyLink: ["sticky-link", toAbs, toRel],
      newsletterActive: ["newsletter-active"],
    });
    return {
      ...data,
      saveData,
    };
  },
};
</script>

<template>
  <details class="mt-4">
    <summary class="title is-4">Sticky popup</summary>
    <BulmaField label="Sticky" help="Pop up is a bottom-right sticky slider">
      <div>
        <label class="checkbox">
          <input v-model="stickyActive" type="checkbox" />
          Show corner sticky to all visitors
        </label>
      </div>
    </BulmaField>
    <template v-if="stickyActive">
      <BulmaFieldInput v-model="stickyLink" label="Sticky link" type="url" />
      <BulmaTextarea
        v-model="stickyImageDescription"
        label="Sticky image description"
        help="For blind readers and search engines"
      />
      <BulmaField
        label="Sticky images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="stickyImages"
          :file-props="fileProps"
          @add="stickyImages.push($event)"
          @remove="stickyImages.splice($event, 1)"
        />
      </BulmaField>
    </template>

    <BulmaField
      label="Newsletter"
      help="Pop up is full screen newsletter takeover"
    >
      <div>
        <label class="checkbox">
          <input v-model="newsletterActive" type="checkbox" />
          Show newsletter pop up
        </label>
      </div>
    </BulmaField>
  </details>
</template>
