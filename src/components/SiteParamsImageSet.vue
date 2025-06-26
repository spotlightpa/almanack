<script setup>
import useProps from "@/utils/use-props.js";

const props = defineProps({
  params: Object,
  propName: String,
  label: String,
  text: String,
  help: String,
  fileProps: Object,
});
let n = 0;

function deserialize(v) {
  let a = v || [];
  return a.map((o) => ({ ...o, id: n++ }));
}

function serialize(v) {
  return v.map((o) => ({ ...o, id: undefined }));
}

let [{ imageSet, active }, saveData] = useProps(props.params.data, {
  imageSet: [props.propName, deserialize, serialize],
  active: [props.propName + "-active", (v) => v ?? false],
});

function pushImage() {
  imageSet.value.push({
    id: n++,
    link: "https://www.spotlightpa.org/donate/",
    description: "Lorem ipsum",
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
    <h4 class="title is-5">
      {{ imageSet.length }} promotion{{ imageSet.length !== 1 ? "s" : "" }}
      active
    </h4>
    <h5 class="subtitle">
      If multiple promotions are active, one will be chosen randomly on page
      load.
    </h5>
    <ul>
      <li v-for="(image, n) of imageSet" :key="image.id" class="zebra-row">
        <BulmaFieldInput
          v-model="image.label"
          label="Banner label"
          placeholder="Sponsored by Acme"
          help="Text accompanying a banner specifying the sponsor or presenter"
        />
        <BulmaFieldInput
          v-model="image.link"
          label="Link URL"
          type="url"
          placeholder="https://www.spotlightpa.org/donate/"
        />
        <BulmaTextarea
          v-model="image.description"
          label="Image description (alt text)"
          help="For blind readers and search engines"
        />
        <BulmaField
          label="Images"
          help="If multiple images are provided, each page load will select one randomly"
        >
          <SiteParamsFiles
            :files="image.sources"
            :file-props="fileProps"
            @add="image.sources.push($event)"
            @remove="image.sources.splice($event, 1)"
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
    <div class="mt-5 buttons">
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
