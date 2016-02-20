#ifndef GO_CLIENT_H
#define GO_CLIENT_H

#include <stdio.h>
#include <pulse/context.h>
#include <pulse/def.h>
#include <pulse/error.h>
#include <pulse/introspect.h>
#include <pulse/thread-mainloop.h>
#include <pulse/stream.h>
#include <pulse/sample.h>

// callback declarations
void            pulse_context_state_callback(pa_context*, void*);
void            pulse_generic_success_callback(pa_context*, int, void*);
void            pulse_generic_index_callback(pa_context*, uint32_t, void*);
void            pulse_get_server_info_callback(pa_context*, const pa_server_info*, void*);
void            pulse_get_sinks_list_callback(pa_context*, const pa_sink_info*, int, void*);
void            pulse_get_sink_info_list_callback(pa_context*, const pa_sink_info*, int, void*);
void            pulse_get_sink_info_by_index_callback(pa_context*, const pa_sink_info*, int, void*);
void            pulse_get_source_info_list_callback(pa_context*, const pa_source_info*, int, void*);
void            pulse_get_source_info_by_index_callback(pa_context*, const pa_source_info*, int, void*);
void            pulse_get_module_info_list_callback(pa_context*, const pa_module_info*, int, void*);
void            pulse_get_module_info_callback(pa_context*, const pa_module_info*, int, void*);
pa_sample_spec  pulse_new_sample_spec(pa_sample_format_t, uint32_t, uint8_t);
void            pulse_stream_success_callback(pa_stream*, int, void*);
void            pulse_stream_state_callback(pa_stream*, void*);
void            pulse_stream_write_callback(pa_stream*, size_t, void*);
int             pulse_stream_write(pa_stream*, void*, size_t, void*);
void            pulse_stream_write_done(void*);
pa_buffer_attr  pulse_stream_get_playback_attr(int32_t, int32_t, int32_t, int32_t);

#endif
