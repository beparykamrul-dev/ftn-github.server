package com.familytimenet.familyguard.core.policy

import android.content.Context
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.familytimenet.familyguard.core.model.FamilyPolicy
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.policyDataStore by preferencesDataStore(name = "ftn_family_guard_policy")

/** Stores the validated local policy document and version; credentials are not persisted here. */
class PolicyRepository(private val context: Context) {
    private val versionKey = longPreferencesKey("policy_version")
    private val jsonKey = stringPreferencesKey("policy_json")

    val policyVersion: Flow<Long> = context.policyDataStore.data.map { it[versionKey] ?: 0L }
    val policyJson: Flow<String?> = context.policyDataStore.data.map { it[jsonKey] }

    suspend fun activate(policy: FamilyPolicy, policyJson: String? = null) {
        require(policy.version > 0) { "Policy version must be positive" }
        context.policyDataStore.updateData { current ->
            current.toMutablePreferences().apply {
                this[versionKey] = policy.version
                if (!policyJson.isNullOrBlank()) this[jsonKey] = policyJson
            }
        }
    }

    suspend fun loadPersisted(): FamilyPolicy? {
        val json = context.policyDataStore.data.map { it[jsonKey] }.let { flow -> flow.firstOrNull() }
        return json?.let { runCatching { PolicyJsonParser.parse(it) }.getOrNull() }
    }

    suspend fun clear() {
        context.policyDataStore.updateData { current ->
            current.toMutablePreferences().apply {
                remove(versionKey)
                remove(jsonKey)
            }
        }
    }

    fun defaultPolicy(): FamilyPolicy = FamilyPolicy.default()
}

private suspend fun <T> kotlinx.coroutines.flow.Flow<T>.firstOrNull(): T? =
    runCatching { kotlinx.coroutines.flow.first(this) }.getOrNull()
