---
title: sleep
type: processor
status: stable
categories: ["Utility"]
---

<!--
     THIS FILE IS AUTOGENERATED!

     To make changes please edit the corresponding source file under internal/impl/<provider>.
-->

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

Sleep for a period of time specified as a duration string for each message. This processor will interpolate functions within the `duration` field, you can find a list of functions [here](/docs/configuration/interpolation#bloblang-queries).

```yml
# Config fields, showing default values
label: ""
sleep:
  duration: "" # No default (required)
```

## Fields

### `duration`

The duration of time to sleep for each execution.
This field supports [interpolation functions](/docs/configuration/interpolation#bloblang-queries).


Type: `string`  


