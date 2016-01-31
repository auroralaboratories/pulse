#include "client.h"

#ifndef FALSE
#define FALSE 0
#endif

#ifndef TRUE
#define TRUE  1
#endif

#define OPROP(op,k,v,t)  go_operationSetProperty(op,k,v,t)
#define OPDONE(op)       go_operationComplete(op)
#define OPERR(op)        go_operationFailed(op)

pa_mainloop     *mainloop = NULL;
pa_context      *context  = NULL;
pa_mainloop_api *api      = NULL;
char            *server   = NULL;


pa_context* pulse_get_context() {
    return context;
}


void pulse_context_state_callback(pa_context *ctx, void *goClient) {
    switch (pa_context_get_state(ctx)) {
    case PA_CONTEXT_CONNECTING:
    case PA_CONTEXT_AUTHORIZING:
    case PA_CONTEXT_SETTING_NAME:
        break;

    case PA_CONTEXT_READY:
        go_clientStartupDone(goClient, "");
        break;

    case PA_CONTEXT_TERMINATED:
        go_clientStartupDone(goClient, "Connection terminated");
        break;
    case PA_CONTEXT_FAILED:
        go_clientStartupDone(goClient, "Connection failed");
        break;
    default:
        go_clientStartupDone(goClient, pa_strerror(pa_context_errno(ctx)));
        break;
    }
}


void pulse_get_server_info_callback(pa_context *ctx, const pa_server_info *info, void *op) {
    char buf[128];

    OPROP(op, "server-string",                   pa_context_get_server(ctx), NULL);
    OPROP(op, "daemon-user",                     info->user_name, NULL);
    OPROP(op, "daemon-hostname",                 info->host_name, NULL);
    OPROP(op, "server-version",                  info->server_version, NULL);
    OPROP(op, "server-name",                     info->server_name, NULL);
    OPROP(op, "default-sink-name",               info->default_sink_name, NULL);
    OPROP(op, "default-source-name",             info->default_source_name, NULL);
    OPROP(op, "sample-format",                   pa_sample_format_to_string(info->sample_spec.format), NULL);

    sprintf(buf, "%d", pa_context_get_server_protocol_version(ctx));
    OPROP(op, "server-protocol-version",         buf, "int");

    sprintf(buf, "%d", pa_context_get_protocol_version(ctx));
    OPROP(op, "library-protocol-version",        buf, "int");

    sprintf(buf, "%d", info->cookie);
    OPROP(op, "cookie",                          buf, "int");

    sprintf(buf, "%d", info->sample_spec.rate);
    OPROP(op, "sample-rate",                     buf, "int");

    sprintf(buf, "%d", info->sample_spec.channels);
    OPROP(op, "channels",                        buf, "int");


    OPDONE(op);
}


int pulse_mainloop_start(const char *name, void *goClient) {
    int code = 0;
    char buffer[64];

    if (!(mainloop = pa_mainloop_new())) {
        go_clientStartupDone(goClient, "Failed to create PulseAudio mainloop");
        return -1;
    }

    api = pa_mainloop_get_api(mainloop);
    context = pa_context_new(api, name);

//  set state change callback for informing the Golang Client{} that we're ready (or have failed)
    pa_context_set_state_callback(context, pulse_context_state_callback, goClient);

//  being context connect
    if (pa_context_connect(context, server, 0, NULL) < 0) {
        go_clientStartupDone(goClient, pa_strerror(pa_context_errno(context)));
        return -1;
    }

//  start pulseaudio mainloop
    if (pa_mainloop_run(mainloop, &code) < 0) {
        sprintf(buffer, "Failed to start mainloop with exit status %d", code);
        go_clientStartupDone(goClient, buffer);
        return -1;
    }

    return code;
}