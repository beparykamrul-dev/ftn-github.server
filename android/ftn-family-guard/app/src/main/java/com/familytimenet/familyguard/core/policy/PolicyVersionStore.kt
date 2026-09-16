package com.familytimenet.familyguard.core.policy

import android.content.Context
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.first

private val Context.policyVersionDataStore by preferencesDataStore(name = "ftn_policy_versions")

/** Prevents accidental rollback to an older policy unless explicitly allowed by control-plane policy. */
class PolicyVersionStore(private val context: Context) {
    private val activeKey = longPreferencesKey("active_version")
    private val previousKey = longPreferencesKey("previous_version")

    suspend fun activeVersion(): Long =
        context.policyVersionDataStore.data.first()[activeKey] ?: 0L

    suspend fun activate(version: Long, allowRollback: Boolean = false): Boolean {
        require(version > 0)
        val current = activeVersion()
        if (current != 0L && version < current && !allowRollback) return false

        context.policyVersionDataStore.updateData { prefs ->
            prefs.toMutablePreferences().apply {
                if (current > 0) this[previousKey] = current
                this[activeKey] = version
            }
        }
        return true
    }
}
