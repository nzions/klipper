// set the printer name and model consts
//
// Copyright (C) 2021  Adrian Keet <arkeet@gmail.com>
//
// This file may be distributed under the terms of the GNU GPLv3 license.

#include "autoconf.h" // CONFIG_PRINTER_NAME CONFIG_PRINTER_MODEL
#include "command.h"  // DECL_CONSTANT_STR

DECL_CONSTANT_STR("machine_name", CONFIG_MACHINE_NAME);
DECL_CONSTANT_STR("machine_model", CONFIG_MACHINE_MODEL);
