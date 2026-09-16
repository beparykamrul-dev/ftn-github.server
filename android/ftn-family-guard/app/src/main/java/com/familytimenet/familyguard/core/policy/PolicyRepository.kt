package com.familytimenet.familyguard.core.policy

import android.content.Context
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.familytimenet.familyguard.core.model.FamilyPolicy
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

private val Context.policyDataStore by preferencesDataStore(name = "ftn_family_guard_policy")

/** Stores only the local policy version/state. Sensitive credentials are not persisted here. */
class PolicyRepository(private val context: Context) {
    private val versionKey = longPreferencesKey("policy_version")

    val policyVersion: Flow<Long> = context.policyDataStore.data.map { it[versionKey] ?: 0L }

    suspend fun activate(policy: FamilyPolicy) {
        require(policy.version > 0) { "Policy version must be positive" }
        context.policyDataStore.updateData { current ->
            current.toMutablePreferences().apply { this[versionKey] = policy.version }
        }
    }

    fun defaultPolicy(): FamilyPolicy = FamilyPolicy.default()
}
