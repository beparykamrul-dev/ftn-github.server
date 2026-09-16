package com.familytimenet.familyguard.core.policy

import com.familytimenet.familyguard.core.model.FamilyPolicy
import com.familytimenet.familyguard.core.realtime.FtnRealtimeClient
import com.familytimenet.familyguard.core.realtime.RealtimeEvent
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import java.util.concurrent.atomic.AtomicLong

/** Coordinates HTTPS policy sync, local validation/cache, acknowledgements and realtime updates. */
class FamilyGuardSyncManager(
    private val policyClient: PolicyClient,
    private val policyRepository: PolicyRepository,
    private val versionStore: PolicyVersionStore,
    private val offlineCache: OfflinePolicyCache,
    private val policyAckClient: PolicyAckClient,
    private val realtimeClient: FtnRealtimeClient,
    private val scope: CoroutineScope
) {
    data class SyncResult(val policy: FamilyPolicy, val source: Source)
    enum class Source { NETWORK, LOCAL, CACHE, DEFAULT }

    private val lastHeartbeat = AtomicLong(0L)

    suspend fun recoverLocal(): SyncResult {
        val persisted = runCatching { policyRepository.loadPersisted() }.getOrNull()
        if (persisted != null && validatePolicy(persisted)) {
            versionStore.activate(persisted.version)
            return SyncResult(persisted, Source.LOCAL)
        }
        val cachedJson = runCatching { offlineCache.load() }.getOrNull()
        if (!cachedJson.isNullOrBlank()) {
            val cached = runCatching { PolicyJsonParser.parse(cachedJson) }.getOrNull()
            if (cached != null && validatePolicy(cached)) {
                policyRepository.activate(cached, cachedJson)
                versionStore.activate(cached.version)
                return SyncResult(cached, Source.CACHE)
            }
        }
        val fallback = policyRepository.defaultPolicy()
        if (validatePolicy(fallback)) return SyncResult(fallback, Source.DEFAULT)
        throw IllegalStateException("No valid Family Guard policy available")
    }

    suspend fun sync(deviceId: String, sessionReference: String): SyncResult {
        val result = runCatching { policyClient.fetchPolicy(deviceId, sessionReference) }.getOrNull()
        val networkPolicy = result?.policy
        if (networkPolicy != null && validatePolicy(networkPolicy)) {
            activate(deviceId, sessionReference, networkPolicy, result.rawJson)
            return SyncResult(networkPolicy, Source.NETWORK)
        }
        return recoverLocal()
    }

    fun startRealtime(deviceId: String, sessionReference: String) = realtimeClient.start(deviceId, sessionReference)
    fun stopRealtime() = realtimeClient.stop()

    fun handleRealtimeEvent(deviceId: String, sessionReference: String, text: String) {
        val event = runCatching { RealtimeEvent.parse(text) }.getOrNull() ?: return
        when (event) {
            is RealtimeEvent.PolicyUpdated -> {
                val policy = runCatching { PolicyJsonParser.parse(event.policyJson) }.getOrNull() ?: return
                if (policy.version != event.version || !validatePolicy(policy)) return
                scope.launch { activate(deviceId, sessionReference, policy, event.policyJson) }
            }
            is RealtimeEvent.PolicyRevoked -> {
                scope.launch {
                    val current = versionStore.activeVersion()
                    if (event.version <= 0L || event.version < current) return@launch
                    offlineCache.clear()
                    val fallback = policyRepository.defaultPolicy()
                    if (validatePolicy(fallback)) {
                        policyRepository.activate(fallback)
                        versionStore.activate(event.version, allowRollback = true)
                    }
                }
            }
            is RealtimeEvent.DeviceCommand -> handleCommand(deviceId, sessionReference, event.command)
            is RealtimeEvent.Heartbeat -> lastHeartbeat.set(event.timestamp)
            is RealtimeEvent.UsageSummary, is RealtimeEvent.Unknown -> Unit
        }
    }

    fun lastHeartbeatTimestamp(): Long = lastHeartbeat.get()

    private fun handleCommand(deviceId: String, sessionReference: String, command: String) {
        when (command.trim().lowercase()) {
            "sync_policy" -> scope.launch { sync(deviceId, sessionReference) }
            "clear_cache" -> scope.launch { offlineCache.clear() }
            "heartbeat" -> realtimeClient.send("{\"type\":\"heartbeat\",\"payload\":{\"timestamp\":${System.currentTimeMillis()}}}")
            "disconnect" -> realtimeClient.stop()
            else -> Unit
        }
    }

    private suspend fun activate(deviceId: String, sessionReference: String, policy: FamilyPolicy, originalJson: String?) {
        val current = versionStore.activeVersion()
        if (current > 0 && policy.version < current) return
        if (!validatePolicy(policy)) return
        policyRepository.activate(policy, originalJson)
        versionStore.activate(policy.version)
        if (!originalJson.isNullOrBlank()) offlineCache.save(originalJson)
        runCatching { policyAckClient.acknowledge(deviceId, sessionReference, policy.version) }
    }

    private fun validatePolicy(policy: FamilyPolicy): Boolean =
        policy.version > 0 && !policy.privacy.rawDnsLogging && !policy.privacy.payloadInspection && policy.privacy.aggregateUsageOnly
}
