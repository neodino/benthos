---
title: to_the_end
type: scanner
status: stable
---

<!--
     THIS FILE IS AUTOGENERATED!

     To make changes please edit the corresponding source file under internal/impl/<provider>.
-->

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

Read the input stream all the way until the end and deliver it as a single message.

```yml
# Config fields, showing default values
to_the_end: {}
```

:::caution
Some sources of data may not have a logical end, therefore caution should be made to exclusively use this scanner when the end of an input stream is clearly defined (and well within memory).
:::



