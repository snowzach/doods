#include <stdlib.h>

#include <darknet.h>

void free_class_names(char **names)
{
    free(names);
}

char ** read_class_names(char *data_cfg)
{
    list *options = read_data_cfg(data_cfg);
    char *name_list = option_find_str(options, "names", "data/names.list");
    return get_labels(name_list);
}

char * get_class_name(char **names, int index, int names_len)
{
    if (index >= names_len) {
        return NULL;
    }

    return names[index];
}
