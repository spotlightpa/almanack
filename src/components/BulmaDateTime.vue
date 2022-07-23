<script setup>
import { computed } from "vue";
import { formatDateTime } from "@/utils/time-format.js";

const props = defineProps({
  label: String,
  help: String,
  value: Date,
  icon: [Array, String],
  required: Boolean,
});

const emit = defineEmits(["input"]);

const leftPad = (n) => String(n).padStart(2, "0");

const localDay = (d) => {
  if (!d) {
    return "";
  }
  let y = leftPad(d.getFullYear());
  let m = leftPad(d.getMonth() + 1);
  let dd = leftPad(d.getDate());
  return `${y}-${m}-${dd}`;
};

const day = computed(() => localDay(props.value));

const time = computed(() => {
  let d = props.value;
  if (!d) {
    return "";
  }
  let h = leftPad(d.getHours());
  let m = leftPad(d.getMinutes());
  let s = leftPad(d.getSeconds());

  return `${h}:${m}:${s}`;
});

function emitDay(value) {
  if (!value) {
    emit("input", null);
    return;
  }
  let t = time.value || "00:00:00";
  emit("input", new Date(`${value}T${t}`));
}

function emitTime(value) {
  if (!day.value && !value) {
    emit("input", null);
    return;
  } else if (!day.value) {
    let today = localDay(new Date());
    emit("input", new Date(`${today}T${value}`));
    return;
  }
  if (!value) {
    value = "00:00:00";
  }
  emit("input", new Date(`${day.value}T${value}`));
}
</script>

<template>
  <BulmaField
    v-slot="{ idForLabel }"
    :label="label"
    :help="help"
    :required="required"
  >
    <div class="my-0 field has-addons">
      <p class="control is-expanded" :class="{ 'has-icons-left': !!icon }">
        <input
          :id="idForLabel"
          class="input"
          type="date"
          :value="day"
          @input="emitDay($event.target.value)"
        />
        <span v-if="icon" class="icon is-left">
          <font-awesome-icon :icon="icon" />
        </span>
      </p>
      <p class="control is-expanded">
        <input
          class="input"
          type="time"
          :value="time"
          @input="emitTime($event.target.value)"
        />
      </p>
      <p class="control">
        <span class="button is-static">
          {{ formatDateTime(value) || "Unset" }}
        </span>
      </p>
    </div>
    <slot />
  </BulmaField>
</template>
