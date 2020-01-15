import Vue from "vue";
import { library } from "@fortawesome/fontawesome-svg-core";
import {
  faCopy,
  faFileCode,
  faFileWord,
  faNewspaper,
} from "@fortawesome/free-regular-svg-icons";
import {
  faFileDownload,
  faLink,
  faFileUpload,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";

library.add(
  faCopy,
  faFileCode,
  faFileWord,
  faNewspaper,
  faFileDownload,
  faLink,
  faFileUpload
);

Vue.component("font-awesome-icon", FontAwesomeIcon);
