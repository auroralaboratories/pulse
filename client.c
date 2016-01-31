#include "client.h"

#ifndef FALSE
#define FALSE 0
#endif

#ifndef TRUE
#define TRUE  1
#endif

#define OPROP(op,k,v,t)  go_operationSetProperty(op,k,v,t)
#define OPINCR(op)       go_operationCreatePayload(op)
#define OPDONE(op)       go_operationComplete(op)
#define OPERR(op,msg)    go_operationFailed(op,msg)


// BAD BAD BAD: this means we only get ONE client instance per process
// TODO: implement more of pulse_mainloop_start() in Golang
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
    char buf[1024];

    OPROP(op, "ServerString",            pa_context_get_server(ctx), NULL);
    OPROP(op, "DaemonUser",              info->user_name, NULL);
    OPROP(op, "DaemonHostname",          info->host_name, NULL);
    OPROP(op, "Version",                 info->server_version, NULL);
    OPROP(op, "Name",                    info->server_name, NULL);
    OPROP(op, "DefaultSinkName",         info->default_sink_name, NULL);
    OPROP(op, "DefaultSourceName",       info->default_source_name, NULL);
    OPROP(op, "SampleFormat",            pa_sample_format_to_string(info->sample_spec.format), NULL);

    sprintf(buf, "%d", pa_context_get_server_protocol_version(ctx));
    OPROP(op, "ProtocolVersion",         buf, "int");

    sprintf(buf, "%d", pa_context_get_protocol_version(ctx));
    OPROP(op, "LibraryProtocolVersion",  buf, "int");

    sprintf(buf, "%d", info->cookie);
    OPROP(op, "Cookie",                  buf, "int");

    sprintf(buf, "%d", info->sample_spec.rate);
    OPROP(op, "SampleRate",              buf, "int");

    sprintf(buf, "%d", info->sample_spec.channels);
    OPROP(op, "Channels",                buf, "int");

    OPDONE(op);
}

void pulse_get_sink_info_list_callback(pa_context *ctx, const pa_sink_info *info, int eol, void *op) {
    char buf[1024];

    if (eol < 0) {
        OPERR(op, pa_strerror(pa_context_errno(ctx)));
    }else{
        if (eol == 0) {

            OPROP(op, "Name",                    info->name, NULL);
            OPROP(op, "Description",             info->description, NULL);
            OPROP(op, "MonitorSourceName",       info->monitor_source_name, NULL);
            OPROP(op, "DriverName",              info->driver, NULL);
            OPROP(op, "Mute",                    (info->mute ? "true" : "false"), "bool");

            sprintf(buf, "%d", info->index);
            OPROP(op, "Index",                   buf, "int");

            sprintf(buf, "%d", info->owner_module);
            OPROP(op, "ModuleIndex",             buf, "int");

            sprintf(buf, "%d", info->monitor_source);
            OPROP(op, "MonitorSourceIndex",      buf, "int");

            sprintf(buf, "%d", info->card);
            OPROP(op, "CardIndex",               buf, "int");

            sprintf(buf, "%d", info->n_ports);
            OPROP(op, "NumPorts",                buf, "int");

            sprintf(buf, "%d", info->n_formats);
            OPROP(op, "NumFormats",              buf, "int");

            sprintf(buf, "%d", info->n_volume_steps);
            OPROP(op, "NumVolumeSteps",          buf, "int");

            sprintf(buf, "%d", info->state);
            OPROP(op, "_state",                  buf, "int");

        //  allocate the next potential response payload
            OPINCR(op);
        }else{
        //  complete the operation; which will resume blocking execution of the Operation.Wait() call
            OPDONE(op);
        }
    }
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