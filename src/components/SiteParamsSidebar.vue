<script>
export default {
  props: { params: Object, fileProps: Object },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Sidebar top</summary>
      <BulmaFieldInput
        v-model="params.sidebarTopLink"
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
          v-model="params.sidebarTopDescription"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Image Width">
          <input
            :id="idForLabel"
            v-model.number="params.sidebarTopWidth"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
        <BulmaField v-slot="{ idForLabel }" class="ml-2" label="Image Height">
          <input
            :id="idForLabel"
            v-model.number="params.sidebarTopHeight"
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
          :files="params.sidebarTopImages"
          :file-props="fileProps"
          @add="params.sidebarTopImages.push($event)"
          @remove="params.sidebarTopImages.splice($event, 1)"
        />
      </BulmaField>
    </details>
    <details class="mt-4">
      <summary class="title is-4">Sidebar sticky</summary>
      <BulmaFieldInput
        v-model="params.sidebarStickyLink"
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
          v-model="params.sidebarStickyDescription"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Image Width">
          <input
            :id="idForLabel"
            v-model.number="params.sidebarStickyWidth"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
        <BulmaField v-slot="{ idForLabel }" class="ml-2" label="Image Height">
          <input
            :id="idForLabel"
            v-model.number="params.sidebarStickyHeight"
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
          :files="params.sidebarStickyImages"
          :file-props="fileProps"
          @add="params.sidebarStickyImages.push($event)"
          @remove="params.sidebarStickyImages.splice($event, 1)"
        />
      </BulmaField>
    </details>
  </div>
</template>
