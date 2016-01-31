#ifndef GO_PULSE_H
#define GO_PULSE_H

#include <stdio.h>
#include <unistd.h>
#include <pulse/context.h>
#include <pulse/error.h>
#include <pulse/introspect.h>
#include <pulse/mainloop.h>
#include <pulse/stream.h>

int         pulse_mainloop_start(void*);
pa_context* pulse_get_context();
void        pulse_get_server_info_callback(pa_context*, const pa_server_info*, void*);
void        pulse_get_sinks_list_callback(pa_context*, const pa_sink_info*, int, void*);


#endif
