#!/bin/bash

cat - | sed -e 's!$FOO!foo!g' -e 's!$BAR!bar!g'
