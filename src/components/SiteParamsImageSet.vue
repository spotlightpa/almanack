<script setup>
import useProps from "@/utils/use-props.js";

const props = defineProps({
  params: Object,
  propName: String,
  label: String,
  text: String,
  help: String,
  fileProps: Object,
  showWidthHeight: Boolean,
});
let n = 0;

function deserialize(v) {
  let a = v || [];
  return a.map((o) => ({
    ...o,
    id: n++,
  }));
}

function serialize(v) {
  return v.map((o) => ({ ...o, id: undefined }));
}

let widthHeightProps = props.showWidthHeight
  ? {
      width: [props.propName + "-width", (v) => v ?? 0],
      height: [props.propName + "-height", (v) => v ?? 0],
    }
  : {};

let [{ imageSet, active, width, height }, saveData] = useProps(
  props.params.data,
  {
    imageSet: [props.propName, deserialize, serialize],
    active: [props.propName + "-active", (v) => v ?? false],
    ...widthHeightProps,
  }
);

function pushImage() {
  imageSet.value.push({
    id: n++,
    label: "",
    labelLink: "",
    link: "https://www.spotlightpa.org/donate/",
    description: "",
    sources: [],
  });
}

function removeImage(n) {
  imageSet.value.splice(n, 1);
}

defineExpose({
  saveData,
});
</script>

<template>
  <BulmaFieldCheckbox v-model="active" :label="label" :help="help">
    {{ " " + text }}
  </BulmaFieldCheckbox>
  <div v-show="active">
    <template v-if="showWidthHeight">
      <p class="help">Set size for promotions</p>
      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Image width">
          <input
            :id="idForLabel"
            v-model.number="width"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
        <BulmaField v-slot="{ idForLabel }" class="ml-2" label="Image height">
          <input
            :id="idForLabel"
            v-model.number="height"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
      </div>
    </template>
    <h4 class="title is-5">
      {{ imageSet.length }} promotion{{ imageSet.length !== 1 ? "s" : "" }}
      active
    </h4>
    <h5 class="subtitle">
      If multiple promotions are active, one will be chosen randomly on page
      load.
    </h5>
    <ul>
      <li v-for="(promo, n) of imageSet" :key="promo.id" class="zebra-row">
        <BulmaFieldInput
          v-model="promo.label"
          label="Banner label"
          placeholder="Sponsored by Acme"
          help="Text accompanying a banner specifying the sponsor or presenter"
        />
        <BulmaFieldInput
          v-model="promo.labelLink"
          label="Banner label link"
          placeholder="https://www.spotlightpa.org/support/"
          help="Link that clicking the ad label will go to"
        />
        <BulmaFieldInput
          v-model="promo.link"
          label="Link URL"
          type="url"
          placeholder="https://www.spotlightpa.org/donate/"
        />
        <BulmaTextarea
          v-model="promo.description"
          label="Image description (alt text)"
          help="For blind readers and search engines"
        />
        <BulmaField
          label="Images"
          help="If multiple images are provided for the same promotion, each page load will select one randomly"
        >
          <SiteParamsFiles
            :files="promo.sources"
            :file-props="fileProps"
            @add="promo.sources.push($event)"
            @remove="promo.sources.splice($event, 1)"
          />
        </BulmaField>
        <div class="mt-1 mb-2 buttons">
          <button
            type="button"
            class="button is-danger has-text-weight-semibold is-small"
            @click="removeImage(n)"
          >
            Remove promotion from set
          </button>
        </div>
      </li>
    </ul>
    <div class="my-5 buttons">
      <button
        type="button"
        class="button is-success has-text-weight-semibold"
        @click="pushImage"
      >
        Add promotion to set
      </button>
    </div>
  </div>
</template>

<style>
.zebra-row {
  background-color: #fff;
}

.zebra-row:nth-child(even) {
  background-color: #fafafa;
}

.zebra-row + .zebra-row {
  border-top: 1px solid #dbdbdb;
}

div {
  box-sizing: border-box;
}
</style>
