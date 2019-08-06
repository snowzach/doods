#pragma once

extern void free_class_names(char **names);
extern char ** read_class_names(char *data_cfg);
extern char * get_class_name(char **names, int index, int names_len);
