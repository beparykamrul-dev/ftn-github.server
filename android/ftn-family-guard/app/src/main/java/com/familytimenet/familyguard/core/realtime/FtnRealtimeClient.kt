package com.familytimenet.familyguard.core.realtime

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import java.net.URI
import java.net.URLEncoder
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean

class FtnRealtimeClient(
    private val socketFactory: SocketFactory,
    private val scope: CoroutineScope,
    private val baseUrl: String
) {
    interface Socket { fun connect(); fun close(); fun send(text:String) }
    interface SocketFactory { fun create(url:String, authorization:String, listener:Listener):Socket }
    interface Listener { fun onOpen(); fun onMessage(text:String); fun onFailure(error:Throwable); fun onClosed() }
    private var socket:Socket?=null
    private var reconnectJob:Job?=null
    private val running=AtomicBoolean(false)
    private var attempt=0

    fun start(deviceId:String,sessionReference:String){
        require(deviceId.isNotBlank());require(sessionReference.isNotBlank())
        val uri=URI.create(baseUrl);require(uri.scheme.equals("wss",true)){"Realtime endpoint must use WSS"}
        running.set(true);attempt=0;connect(deviceId,sessionReference)
    }
    fun stop(){running.set(false);reconnectJob?.cancel();reconnectJob=null;socket?.close();socket=null}
    fun send(text:String){if(running.get()&&text.isNotBlank())socket?.send(text)}
    private fun connect(deviceId:String,session:String){
        if(!running.get())return
        socket?.close()
        socket=socketFactory.create(buildUrl(deviceId),"Bearer $session",object:Listener{
            override fun onOpen(){attempt=0}
            override fun onMessage(text:String){/* Routed by sync manager. */}
            override fun onFailure(error:Throwable){scheduleReconnect(deviceId,session)}
            override fun onClosed(){scheduleReconnect(deviceId,session)}
        })
        socket?.connect()
    }
    private fun scheduleReconnect(deviceId:String,session:String){
        if(!running.get()||reconnectJob?.isActive==true)return
        reconnectJob=scope.launch{val exponent=attempt.coerceAtMost(6);val wait=(1000L shl exponent).coerceAtMost(60000L);attempt++;delay(wait);if(isActive&&running.get())connect(deviceId,session)}
    }
    private fun buildUrl(deviceId:String)=baseUrl.trimEnd('/')+"/ws/family?device_id="+URLEncoder.encode(deviceId,"UTF-8").replace("+","%20")
}

class OkHttpSocketFactory(private val client:OkHttpClient=OkHttpClient.Builder().pingInterval(30,TimeUnit.SECONDS).build()):FtnRealtimeClient.SocketFactory{
    override fun create(url:String,authorization:String,listener:FtnRealtimeClient.Listener):FtnRealtimeClient.Socket=object:FtnRealtimeClient.Socket{
        var ws:WebSocket?=null
        override fun connect(){ws=client.newWebSocket(Request.Builder().url(url).header("Authorization",authorization).build(),object:WebSocketListener(){
            override fun onOpen(webSocket:WebSocket,response:Response){listener.onOpen()}
            override fun onMessage(webSocket:WebSocket,text:String){listener.onMessage(text)}
            override fun onFailure(webSocket:WebSocket,t:Throwable,response:Response?){listener.onFailure(t)}
            override fun onClosed(webSocket:WebSocket,code:Int,reason:String){listener.onClosed()}
        })}
        override fun close(){ws?.close(1000,"closed");ws=null}
        override fun send(text:String){ws?.send(text)}
    }
}
