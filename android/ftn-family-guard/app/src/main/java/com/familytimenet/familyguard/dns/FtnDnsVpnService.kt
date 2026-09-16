package com.familytimenet.familyguard.dns

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Intent
import android.net.VpnService
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat
import com.familytimenet.familyguard.core.filter.RuleEngine
import com.familytimenet.familyguard.core.model.FamilyPolicy

/**
 * User-visible VPN shell for DNS protection.
 *
 * This service intentionally does not inspect application payloads. A complete
 * packet/TUN forwarding implementation must be added as a separate, tested
 * transport module before production traffic is routed through the VPN.
 */
class FtnDnsVpnService : VpnService() {
    private val ruleEngine = RuleEngine { FamilyPolicy.default() }
    private var tunnel: android.os.ParcelFileDescriptor? = null

    override fun onCreate() {
        super.onCreate()
        createNotificationChannel()
        startForeground(NOTIFICATION_ID, notification())
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        if (tunnel == null) {
            tunnel = Builder()
                .setSession("FTN Family Guard")
                .addAddress("10.245.0.2", 32)
                .addRoute("10.245.0.0", 24)
                .establish()
        }
        return START_STICKY
    }

    override fun onDestroy() {
        tunnel?.close()
        tunnel = null
        super.onDestroy()
    }

    override fun onBind(intent: Intent): IBinder? = super.onBind(intent)

    private fun notification(): android.app.Notification =
        NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("FTN Family Guard")
            .setContentText("DNS protection is active")
            .setSmallIcon(android.R.drawable.stat_sys_warning)
            .setOngoing(true)
            .build()

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(
                NotificationChannel(
                    CHANNEL_ID,
                    "FTN Family Guard",
                    NotificationManager.IMPORTANCE_LOW
                )
            )
        }
    }

    companion object {
        private const val CHANNEL_ID = "ftn-family-guard"
        private const val NOTIFICATION_ID = 4102
    }
}
