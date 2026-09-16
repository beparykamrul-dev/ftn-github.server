package com.familytimenet.familyguard.core.auth

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import java.nio.charset.StandardCharsets
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

private val Context.sessionDataStore by preferencesDataStore(name = "ftn_secure_session")

/** Encrypts opaque session references with an Android Keystore AES-256-GCM key before persistence. */
class SecureSessionStore(private val context: Context) {
    private val sessionKey = stringPreferencesKey("session_reference")
    private val enrollmentKey = stringPreferencesKey("enrollment_id")
    private val alias = "ftn_family_guard_session_v1"

    val sessionReference: Flow<String?> =
        context.sessionDataStore.data.map { it[sessionKey]?.let(::decrypt) }

    val enrollmentId: Flow<String?> =
        context.sessionDataStore.data.map { it[enrollmentKey]?.let(::decrypt) }

    suspend fun save(sessionReference: String, enrollmentId: String) {
        require(sessionReference.isNotBlank())
        require(enrollmentId.isNotBlank())
        context.sessionDataStore.updateData { current ->
            current.toMutablePreferences().apply {
                this[sessionKey] = encrypt(sessionReference)
                this[enrollmentKey] = encrypt(enrollmentId)
            }
        }
    }

    suspend fun clear() {
        context.sessionDataStore.updateData { current ->
            current.toMutablePreferences().apply {
                remove(sessionKey)
                remove(enrollmentKey)
            }
        }
    }

    private fun key(): SecretKey {
        val store = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
        (store.getKey(alias, null) as? SecretKey)?.let { return it }
        val generator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore")
        generator.init(
            KeyGenParameterSpec.Builder(
                alias,
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
            )
                .setKeySize(256)
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setUserAuthenticationRequired(false)
                .build()
        )
        return generator.generateKey()
    }

    private fun encrypt(value: String): String {
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, key())
        val encrypted = cipher.doFinal(value.toByteArray(StandardCharsets.UTF_8))
        return b64(cipher.iv) + "." + b64(encrypted)
    }

    private fun decrypt(value: String): String? = runCatching {
        val parts = value.split('.', limit = 2)
        require(parts.size == 2)
        val iv = android.util.Base64.decode(parts[0], android.util.Base64.NO_WRAP)
        val data = android.util.Base64.decode(parts[1], android.util.Base64.NO_WRAP)
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.DECRYPT_MODE, key(), GCMParameterSpec(128, iv))
        String(cipher.doFinal(data), StandardCharsets.UTF_8)
    }.getOrNull()

    private fun b64(value: ByteArray): String =
        android.util.Base64.encodeToString(value, android.util.Base64.NO_WRAP)
}
