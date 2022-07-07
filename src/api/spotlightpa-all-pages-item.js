export default class AllPagesItem {
  constructor(data) {
    this.id = data.id;
    this.filePath = data.file_path ?? "";
    this.internalID = data.internal_id ?? "";
    this.hed = data.hed ?? "";
    this.authors = data.authors ?? [];
    this.filterableProps = `${this.internalID} ${this.hed} ${this.authors.join(
      " "
    )}`;
  }
}
