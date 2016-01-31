#ifndef GO_CLIENT_H
#define GO_CLIENT_H

#include <stdio.h>
#include <pulse/context.h>
#include <pulse/def.h>
#include <pulse/error.h>
#include <pulse/introspect.h>
#include <pulse/mainloop.h>
#include <pulse/stream.h>
#include <pulse/sample.h>

int         pulse_mainloop_start(const char*, void*);
pa_context* pulse_get_context();

// callback declarations
void        pulse_generic_success_callback(pa_context*, int, void*);
void        pulse_get_server_info_callback(pa_context*, const pa_server_info*, void*);
void        pulse_get_sinks_list_callback(pa_context*, const pa_sink_info*, int, void*);
void        pulse_get_sink_info_list_callback(pa_context*, const pa_sink_info*, int, void*);
void        pulse_get_sink_info_by_index_callback(pa_context*, const pa_sink_info*, int, void*);

#endif
