# Release verification keys

This directory contains all keys that are currently signing Alby Hub releases.

The name of the file must match exactly the suffix that user is going to use
when signing a release.
For example, if the key is called `im-adithya.asc` then that user should upload a
signature file called `manifest-im-adithya.txt.asc`.

In addition to adding the key file here there is a main `mainfest.txt.asc` file
that is used by the systemd install/update scripts. This must be provided by a
single verified signing user per release.
