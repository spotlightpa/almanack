import { ref, type Ref, watch } from "vue";

import { get, post, getSiteData, postSiteData } from "@/api/client.ts";
import { makeState } from "@/api/loader.ts";
// @ts-expect-error untyped JS module
import { useFileList } from "@/api/file-list.js";
// @ts-expect-error untyped JS module
import useScrollTo from "@/utils/use-scroll-to.js";
import maybeDate from "@/utils/maybe-date.ts";

class SiteParamsModel {
  scheduleFor: Date | null;
  publishedAt: Date | null;
  isCurrent: boolean;
  data: Record<string, unknown>;

  constructor(config: Record<string, unknown>) {
    this.scheduleFor = maybeDate(config, "schedule_for");
    this.publishedAt = maybeDate(config, "published_at");
    this.isCurrent = !!this.publishedAt;
    this.data = (config.data ?? {}) as Record<string, unknown>;
  }

  toJSON() {
    return {
      schedule_for: this.scheduleFor,
      data: this.data,
    };
  }
}

export function useSiteParamsEditor(location: string) {
  const query = `?location=${location}`;

  const scheduledConfigs = ref<SiteParamsModel[]>([]);
  const siteParamsComps = ref<{ saveParams(): unknown }[]>([]);
  const nextSchedule = ref<Date | null>(null);

  const { exec, apiStateRefs } = makeState();

  function fetch() {
    return exec(() => get(getSiteData + query));
  }

  const [container, scrollTo] = useScrollTo();

  async function addScheduledConfig() {
    let lastParams = scheduledConfigs.value[
      scheduledConfigs.value.length - 1
    ] ?? { data: {} };
    scheduledConfigs.value.push(
      new SiteParamsModel({
        ...JSON.parse(JSON.stringify(lastParams)),
        schedule_for: nextSchedule.value,
      })
    );
    nextSchedule.value = null;
    await scrollTo();
  }

  function removeScheduledConfig(i: number) {
    scheduledConfigs.value.splice(i, 1);
  }

  async function save() {
    let configs = siteParamsComps.value.map((comp) => comp.saveParams());
    return exec(() => post(postSiteData + query, { configs }));
  }

  watch(apiStateRefs.rawData as Ref<Record<string, unknown>>, (data) => {
    if (!data?.configs) {
      return;
    }
    scheduledConfigs.value = (data.configs as Record<string, unknown>[]).map(
      (cfg) => new SiteParamsModel(cfg)
    );
  });

  const { isLoading, isLoadingThrottled, error } = apiStateRefs;

  const files = useFileList();

  fetch();

  return {
    scheduledConfigs,
    siteParamsComps,
    nextSchedule,
    container,
    isLoading,
    isLoadingThrottled,
    error,
    files,
    fetch,
    save,
    addScheduledConfig,
    removeScheduledConfig,
  };
}
