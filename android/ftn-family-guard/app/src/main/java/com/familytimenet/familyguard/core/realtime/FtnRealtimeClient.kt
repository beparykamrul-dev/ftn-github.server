package com.familytimenet.familyguard.core.realtime

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import java.net.HttpURLConnection
import java.net.URI
import java.util.concurrent.atomic.AtomicBoolean

/**
 * Realtime lifecycle contract for the Android Family Guard client.
 * The concrete socket transport is injected later, keeping Android policy logic
 * independent from a particular WebSocket library.
 */
class FtnRealtimeClient(
    private val socketFactory: SocketFactory,
    private val scope: CoroutineScope,
    private val baseUrl: String
) {
    interface Socket {
        fun connect()
        fun close()
        fun send(text: String)
    }

    interface SocketFactory {
        fun create(url: String, authorization: String, listener: Listener): Socket
    }

    interface Listener {
        fun onOpen()
        fun onMessage(text: String)
        fun onFailure(error: Throwable)
        fun onClosed()
    }

    private var socket: Socket? = null
    private var reconnectJob: Job? = null
    private val running = AtomicBoolean(false)
    private var attempt = 0

    fun start(deviceId: String, sessionReference: String) {
        require(deviceId.isNotBlank())
        require(sessionReference.isNotBlank())
        val uri = URI.create(baseUrl)
        require(uri.scheme.equals("wss", ignoreCase = true)) { "Realtime endpoint must use WSS" }
        running.set(true)
        attempt = 0
        connect(deviceId, sessionReference)
    }

    fun stop() {
        running.set(false)
        reconnectJob?.cancel()
        reconnectJob = null
        socket?.close()
        socket = null
    }

    fun send(text: String) {
        if (running.get() && text.isNotBlank()) socket?.send(text)
    }

    private fun connect(deviceId: String, sessionReference: String) {
        if (!running.get()) return
        socket?.close()
        socket = socketFactory.create(
            buildUrl(deviceId),
            "Bearer $sessionReference",
            object : Listener {
                override fun onOpen() {
                    attempt = 0
                }

                override fun onMessage(text: String) {
                    // Event routing is intentionally handled by the policy/usage layers.
                }

                override fun onFailure(error: Throwable) {
                    scheduleReconnect(deviceId, sessionReference)
                }

                override fun onClosed() {
                    scheduleReconnect(deviceId, sessionReference)
                }
            }
        )
        socket?.connect()
    }

    private fun scheduleReconnect(deviceId: String, sessionReference: String) {
        if (!running.get() || reconnectJob?.isActive == true) return
        reconnectJob = scope.launch {
            val exponent = attempt.coerceAtMost(6)
            val waitMs = (1_000L shl exponent).coerceAtMost(60_000L)
            attempt++
            delay(waitMs)
            if (isActive && running.get()) connect(deviceId, sessionReference)
        }
    }

    private fun buildUrl(deviceId: String): String =
        baseUrl.trimEnd('/') + "/api/v1/family/device/stream?device_id=" +
            java.net.URLEncoder.encode(deviceId, "UTF-8").replace("+", "%20")
}
