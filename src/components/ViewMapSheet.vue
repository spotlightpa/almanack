<template>
  <MetaHead>
    <title>Import Custom Template • Spotlight PA Almanack</title>
  </MetaHead>

  <div class="px-2">
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Import Custom Template', to: { name: 'map-sheet' } },
      ]"
    ></BulmaBreadcrumbs>
    <h1 class="title">Import Custom Template</h1>
  </div>

  <MapSheetImporter></MapSheetImporter>
</template>
