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
    <div v-show="stickyActive">
      <BulmaFieldInput
        v-model="stickyLink"
        label="Sticky link"
        type="url"
      ></BulmaFieldInput>
      <BulmaTextarea
        v-model="stickyImageDescription"
        label="Sticky image description"
        help="For blind readers and search engines"
      ></BulmaTextarea>
      <BulmaField
        label="Sticky images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="stickyImages"
          :file-props="fileProps"
          @add="stickyImages.push($event)"
          @remove="stickyImages.splice($event, 1)"
        ></SiteParamsFiles>
      </BulmaField>
    </div>
  </details>
</template>
