import Vue from "vue";

import config from "buefy/src/utils/config";
import {
  Autocomplete,
  Datetimepicker,
  Field,
  Input,
  Numberinput,
  Select,
  Taginput,
} from "buefy/src/components";

const MyBuefy = {
  install(v, options = {}) {
    Object.assign(config, options, {
      defaultIconComponent: "font-awesome-icon",
      defaultIconPack: "fas",
    });

    let comps = [
      Autocomplete,
      Datetimepicker,
      Field,
      Input,
      Numberinput,
      Select,
      Taginput,
    ];

    for (let comp of comps) {
      v.use(comp);
    }
  },
};

Vue.use(MyBuefy);
