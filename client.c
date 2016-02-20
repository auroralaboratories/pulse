#include "client.h"

#ifndef FALSE
#define FALSE 0
#endif

#ifndef TRUE
#define TRUE  1
#endif

// set a named property (k) to the value (v) with type hint (t) within the current operation payload (op)
#define OPROP(op,k,v,t)  go_operationSetProperty(op,k,v,t)

// create a new payload within the given operation (op)
#define OPINCR(op)       go_operationCreatePayload(op)

// finalize the current operation (op) [successful]
#define OPDONE(op)       go_operationComplete(op)

// fail the current operation (op) with a given error message (msg)
#define OPERR(op,msg)    go_operationFailed(op,msg)



// macros for configuring how various values are formatted for Golang
//
#define SINK_VOLUME_AGGREGATOR(v)      pa_cvolume_avg(v)
#define SINK_VOLUME_FACTOR_PRECISION   "4"
#define SOURCE_VOLUME_AGGREGATOR(v)    pa_cvolume_avg(v)
#define SOURCE_VOLUME_FACTOR_PRECISION "4"


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

void pulse_generic_success_callback(pa_context *ctx, int success, void *op) {
    if (success) {
        OPDONE(op);
    }else{
        OPERR(op, pa_strerror(pa_context_errno(ctx)));
    }
}

void pulse_generic_index_callback(pa_context *ctx, uint32_t index, void *op) {
    char buf[1024];

    if (index != PA_INVALID_INDEX) {
        sprintf(buf, "%d", index);
        OPROP(op, "Index", buf, "int");
        OPDONE(op);
    }else{
        OPERR(op, pa_strerror(pa_context_errno(ctx)));
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

void pulse_get_sink_info_by_index_callback(pa_context *ctx, const pa_sink_info *info, int eol, void *op) {
    char buf[1024];

    if (eol < 0) {
        OPERR(op, pa_strerror(pa_context_errno(ctx)));
    }else{
        if (eol == 0) {
            OPROP(op, "Name",                    info->name, NULL);
            OPROP(op, "Description",             info->description, NULL);
            OPROP(op, "MonitorSourceName",       info->monitor_source_name, NULL);
            OPROP(op, "DriverName",              info->driver, NULL);
            OPROP(op, "Muted",                   (info->mute ? "true" : "false"), "bool");

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


        //  volume retrieval and parsing
            if(info->volume.channels > 0){
                sprintf(buf, "%d", info->volume.channels);
                OPROP(op, "Channels",            buf, "int");

                pa_volume_t aggregateVolume;
                pa_volume_t min = pa_cvolume_min(&info->volume);
                pa_volume_t max = pa_cvolume_max(&info->volume);

                if(min == max){
                    aggregateVolume = max;
                }else{
                    aggregateVolume = SINK_VOLUME_AGGREGATOR(&info->volume);
                }

                sprintf(buf, "%d", aggregateVolume);
                OPROP(op, "CurrentVolumeStep",   buf, "int");

                sprintf(buf, "%."SINK_VOLUME_FACTOR_PRECISION"f", ((double)aggregateVolume / (double)info->n_volume_steps));
                OPROP(op, "VolumeFactor",        buf, "float");
            }

        //  allocate the next potential response payload
            OPINCR(op);
        }else{
        //  complete the operation; which will resume blocking execution of the Operation.Wait() call
            OPDONE(op);
        }
    }
}

void pulse_get_sink_info_list_callback(pa_context *ctx, const pa_sink_info *info, int eol, void *op) {
    if (eol < 0) {
        OPERR(op, pa_strerror(pa_context_errno(ctx)));
    }else{
    //  use the ..sink_info_by_index callback to reuse the same logic that Sink.Refresh() uses without
    //  doing the call twice
        pulse_get_sink_info_by_index_callback(ctx, info, eol, op);
    }
}

void pulse_get_source_info_by_index_callback(pa_context *ctx, const pa_source_info *info, int eol, void *op) {
    char buf[1024];

    if (eol < 0) {
        OPERR(op, pa_strerror(pa_context_errno(ctx)));
    }else{
        if (eol == 0) {
            OPROP(op, "Name",                    info->name, NULL);
            OPROP(op, "Description",             info->description, NULL);
            OPROP(op, "DriverName",              info->driver, NULL);
            OPROP(op, "MonitorOfSinkName",       info->monitor_of_sink_name, NULL);
            OPROP(op, "Muted",                   (info->mute ? "true" : "false"), "bool");

            sprintf(buf, "%d", info->index);
            OPROP(op, "Index",                   buf, "int");

            sprintf(buf, "%d", info->owner_module);
            OPROP(op, "ModuleIndex",             buf, "int");

            sprintf(buf, "%d", info->monitor_of_sink);
            OPROP(op, "MonitorOfSinkIndex",      buf, "int");

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


        //  volume retrieval and parsing
            if(info->volume.channels > 0){
                sprintf(buf, "%d", info->volume.channels);
                OPROP(op, "Channels",            buf, "int");

                pa_volume_t aggregateVolume;
                pa_volume_t min = pa_cvolume_min(&info->volume);
                pa_volume_t max = pa_cvolume_max(&info->volume);

                if(min == max){
                    aggregateVolume = max;
                }else{
                    aggregateVolume = SOURCE_VOLUME_AGGREGATOR(&info->volume);
                }

                sprintf(buf, "%d", aggregateVolume);
                OPROP(op, "CurrentVolumeStep",   buf, "int");

                sprintf(buf, "%."SOURCE_VOLUME_FACTOR_PRECISION"f", ((double)aggregateVolume / (double)info->n_volume_steps));
                OPROP(op, "VolumeFactor",        buf, "float");
            }

        //  allocate the next potential response payload
            OPINCR(op);
        }else{
        //  complete the operation; which will resume blocking execution of the Operation.Wait() call
            OPDONE(op);
        }
    }
}

void pulse_get_source_info_list_callback(pa_context *ctx, const pa_source_info *info, int eol, void *op) {
    if (eol < 0) {
        OPERR(op, pa_strerror(pa_context_errno(ctx)));
    }else{
        pulse_get_source_info_by_index_callback(ctx, info, eol, op);
    }
}



void pulse_get_module_info_callback(pa_context *ctx, const pa_module_info *info, int eol, void *op) {
    char buf[1024];

    if (eol < 0) {
        OPERR(op, pa_strerror(pa_context_errno(ctx)));
    }else{
        if (eol == 0) {
            OPROP(op, "Name",                    info->name, NULL);
            OPROP(op, "Argument",                info->argument, NULL);

            sprintf(buf, "%d", info->index);
            OPROP(op, "Index",                   buf, "int");

            sprintf(buf, "%d", info->n_used);
            OPROP(op, "NumUsed",                 buf, "int");


        //  allocate the next potential response payload
            OPINCR(op);
        }else{
        //  complete the operation; which will resume blocking execution of the Operation.Wait() call
            OPDONE(op);
        }
    }
}

void pulse_get_module_info_list_callback(pa_context *ctx, const pa_module_info *info, int eol, void *op) {
    if (eol < 0) {
        OPERR(op, pa_strerror(pa_context_errno(ctx)));
    }else{
        pulse_get_module_info_callback(ctx, info, eol, op);
    }
}

pa_sample_spec pulse_new_sample_spec(pa_sample_format_t fmt, uint32_t rt, uint8_t nchan) {
    pa_sample_spec ss = {
        .format   = fmt,
        .rate     = rt,
        .channels = nchan
    };

    return ss;
}

void pulse_stream_state_callback(pa_stream *stream, void *goStream) {
    pa_context *ctx = pa_stream_get_context(stream);

    switch (pa_stream_get_state(stream)) {
    case PA_STREAM_UNCONNECTED:
    case PA_STREAM_CREATING:
        break;

    case PA_STREAM_READY:
        go_streamStateChange(goStream, "");
        break;

    case PA_STREAM_FAILED:
        go_streamStateChange(goStream, pa_strerror(pa_context_errno(ctx)));
        break;

    case PA_STREAM_TERMINATED:
        go_streamStateChange(goStream, "Connection terminated");
        break;

    default:
        go_streamStateChange(goStream, pa_strerror(pa_context_errno(ctx)));
        break;

    }
}

// this callback will inform the proper stream that it's time to perform write
// of size `len' from it's internal buffer
//
void pulse_stream_write_callback(pa_stream *stream, size_t len, void *op) {
    go_streamPerformWrite(op, len);
}


void pulse_stream_success_callback(pa_stream *stream, int success, void *op) {
    if(success > 0){
        OPDONE(op);
    }else{
        pa_context *ctx = pa_stream_get_context(stream);
        char buf[128];

        sprintf(buf, "Stream operation failed: %s", pa_strerror(pa_context_errno(ctx)));
        OPERR(op, buf);
    }
}

int pulse_stream_write(pa_stream *stream, void *data, size_t len, void *op) {
    pa_context *ctx = pa_stream_get_context(stream);
    char buf[64];
    size_t actual = len;

    int status = pa_stream_begin_write(stream, &data, &actual);

    if(status < 0) {
        return status;
    }else{
        status = pa_stream_write(stream, data, actual, NULL, 0, PA_SEEK_RELATIVE);

        if(status < 0) {
            return status;
        }else{
            return (int)(actual);
        }
    }
}


pa_buffer_attr pulse_stream_get_playback_attr(int32_t ml, int32_t tlen, int32_t pb, int32_t mreq) {
    pa_buffer_attr buffer_attr;

    buffer_attr.maxlength = (uint32_t)(ml);
    buffer_attr.tlength   = (uint32_t)(tlen);
    buffer_attr.prebuf    = (uint32_t)(pb);
    buffer_attr.minreq    = (uint32_t)(mreq);

    return buffer_attr;
}
