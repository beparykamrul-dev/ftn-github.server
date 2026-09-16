package com.familytimenet.familyguard.core.policy

import android.content.Context
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.first

private val Context.offlinePolicyDataStore by preferencesDataStore(name = "ftn_family_guard_offline_policy")

/** Keeps the last validated policy locally so filtering can continue while offline. */
class OfflinePolicyCache(private val context: Context) {
    private val jsonKey = stringPreferencesKey("active_policy_json")

    suspend fun save(policyJson: String) {
        require(policyJson.isNotBlank())
        context.offlinePolicyDataStore.updateData { prefs ->
            prefs.toMutablePreferences().apply { this[jsonKey] = policyJson }
        }
    }

    suspend fun load(): String? =
        context.offlinePolicyDataStore.data.first()[jsonKey]

    suspend fun clear() {
        context.offlinePolicyDataStore.updateData { prefs ->
            prefs.toMutablePreferences().apply { remove(jsonKey) }
        }
    }
}
