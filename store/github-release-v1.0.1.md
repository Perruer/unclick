A fix release for the step that makes generated configurations pass provider validation. Thanks to the reader on Habr whose questions found the problem.

## Fixed

- Validation fixes no longer stop after 16 rounds and leave a resource invalid without a word. The loop now runs until every optional argument the provider rejects is gone; the number of arguments present bounds it, so it always ends.
- When the provider rejects several arguments, those holding a zero value are left out first: the legacy plugin SDK writes zero values for arguments that were never set.

## Added

- Generated files say what was changed: arguments left out, and provider errors that remain, are written as comments above the resource. The log ends with a list of resources that still fail validation.
- A test that fails when the plugin protocol files gain a field the protocol 5 and 6 adapters neither map nor skip on purpose.

Binaries for Linux, macOS and Windows are attached below, with `checksums.txt`.

Full list: [CHANGELOG](https://github.com/Perruer/unclick/blob/main/CHANGELOG.md).
