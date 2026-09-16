package com.familytimenet.familyguard.core.policy

import com.familytimenet.familyguard.core.model.FamilyPolicy
import com.familytimenet.familyguard.core.realtime.FtnRealtimeClient
import com.familytimenet.familyguard.core.realtime.RealtimeEvent
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.launch

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
    enum class Source { NETWORK, CACHE, DEFAULT }

    suspend fun sync(deviceId: String, sessionReference: String): SyncResult {
        val result = runCatching { policyClient.fetchPolicy(deviceId, sessionReference) }.getOrNull()
        val networkPolicy = result?.policy
        if (networkPolicy != null && validatePolicy(networkPolicy)) {
            activate(deviceId, sessionReference, networkPolicy, result.rawJson)
            return SyncResult(networkPolicy, Source.NETWORK)
        }

        val cached = runCatching { offlineCache.load() }.getOrNull()
        if (!cached.isNullOrBlank()) {
            val policy = runCatching { PolicyJsonParser.parse(cached) }.getOrNull()
            if (policy != null && validatePolicy(policy)) {
                policyRepository.activate(policy)
                versionStore.activate(policy.version)
                return SyncResult(policy, Source.CACHE)
            }
        }

        return SyncResult(policyRepository.defaultPolicy(), Source.DEFAULT)
    }

    fun startRealtime(deviceId: String, sessionReference: String) =
        realtimeClient.start(deviceId, sessionReference)

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
                    if (event.version > 0 && event.version >= versionStore.activeVersion()) {
                        offlineCache.clear()
                    }
                }
            }
            is RealtimeEvent.DeviceCommand -> Unit
            is RealtimeEvent.Heartbeat,
            is RealtimeEvent.UsageSummary,
            is RealtimeEvent.Unknown -> Unit
        }
    }

    private suspend fun activate(
        deviceId: String,
        sessionReference: String,
        policy: FamilyPolicy,
        originalJson: String?
    ) {
        val current = versionStore.activeVersion()
        if (current > 0 && policy.version < current) return

        policyRepository.activate(policy)
        versionStore.activate(policy.version)
        if (!originalJson.isNullOrBlank()) offlineCache.save(originalJson)
        runCatching { policyAckClient.acknowledge(deviceId, sessionReference, policy.version) }
    }

    private fun validatePolicy(policy: FamilyPolicy): Boolean =
        policy.version > 0 &&
            !policy.privacy.rawDnsLogging &&
            !policy.privacy.payloadInspection &&
            policy.privacy.aggregateUsageOnly
}
