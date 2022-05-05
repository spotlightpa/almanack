<script>
import useData from "@/utils/use-data.js";
import { toRel, toAbs } from "@/utils/link.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    return {
      ...useData(() => props.params.data, {
        sidebarTopDescription: ["sidebar-top-description"],
        sidebarTopLink: ["sidebar-top-link", toAbs, toRel],
        sidebarTopImages: ["sidebar-top-images"],
        sidebarTopWidth: ["sidebar-top-width"],
        sidebarTopHeight: ["sidebar-top-height"],

        sidebarStickyDescription: ["sidebar-sticky-description"],
        sidebarStickyLink: ["sidebar-sticky-link", toAbs, toRel],
        sidebarStickyImages: ["sidebar-sticky-images"],
        sidebarStickyWidth: ["sidebar-sticky-width"],
        sidebarStickyHeight: ["sidebar-sticky-height"],
      }),
    };
  },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Sidebar top</summary>
      <BulmaFieldInput
        v-model="sidebarTopLink"
        label="Sidebar top promo link"
        type="url"
      />
      <BulmaField
        v-slot="{ idForLabel }"
        label="Sidebar top promo image description"
        help="For blind readers and search engines"
      >
        <textarea
          :id="idForLabel"
          v-model="sidebarTopDescription"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Image Width">
          <input
            :id="idForLabel"
            v-model.number="sidebarTopWidth"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
        <BulmaField v-slot="{ idForLabel }" class="ml-2" label="Image Height">
          <input
            :id="idForLabel"
            v-model.number="sidebarTopHeight"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
      </div>
      <BulmaField
        label="Images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="sidebarTopImages"
          :file-props="fileProps"
          @add="sidebarTopImages.push($event)"
          @remove="sidebarTopImages.splice($event, 1)"
        />
      </BulmaField>
    </details>
    <details class="mt-4">
      <summary class="title is-4">Sidebar sticky</summary>
      <BulmaFieldInput
        v-model="sidebarStickyLink"
        label="Sidebar sticky bottom promo link"
        type="url"
      />
      <BulmaField
        v-slot="{ idForLabel }"
        label="Sidebar sticky bottom promo image description"
        help="For blind readers and search engines"
      >
        <textarea
          :id="idForLabel"
          v-model="sidebarStickyDescription"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Image Width">
          <input
            :id="idForLabel"
            v-model.number="sidebarStickyWidth"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
        <BulmaField v-slot="{ idForLabel }" class="ml-2" label="Image Height">
          <input
            :id="idForLabel"
            v-model.number="sidebarStickyHeight"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
      </div>
      <BulmaField
        label="Images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="sidebarStickyImages"
          :file-props="fileProps"
          @add="sidebarStickyImages.push($event)"
          @remove="sidebarStickyImages.splice($event, 1)"
        />
      </BulmaField>
    </details>
  </div>
</template>
