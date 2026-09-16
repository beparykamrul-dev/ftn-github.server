package com.familytimenet.familyguard.dns

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Intent
import android.net.VpnService
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat
import com.familytimenet.familyguard.core.filter.FamilyGuardFilter
import com.familytimenet.familyguard.core.model.FamilyPolicy
import com.familytimenet.familyguard.core.usage.UsageAggregator

/**
 * User-visible VPN shell for Family Guard.
 *
 * The service owns the local TUN interface and policy decision layer. It does not
 * inspect application payloads or message contents. Full DNS packet forwarding is
 * intentionally kept separate until a tested transport implementation is available.
 */
class FtnDnsVpnService : VpnService() {
    private var tunnel: android.os.ParcelFileDescriptor? = null
    private var activePolicy: FamilyPolicy = FamilyPolicy.default()
    private val filter = FamilyGuardFilter { activePolicy }
    private val usage = UsageAggregator()

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

    fun activatePolicy(policy: FamilyPolicy) {
        require(policy.version > 0)
        require(!policy.privacy.rawDnsLogging)
        require(!policy.privacy.payloadInspection)
        activePolicy = policy
    }

    fun evaluateDomain(domain: String) = filter.evaluateDomain(domain).also {
        if (it.action == com.familytimenet.familyguard.core.filter.FilterDecision.Action.ALLOW) {
            usage.recordAllowed()
        } else {
            usage.recordBlocked()
        }
    }

    fun evaluateIpv4(address: String) = filter.evaluateIpv4(address).also {
        if (it.action == com.familytimenet.familyguard.core.filter.FilterDecision.Action.ALLOW) {
            usage.recordAllowed()
        } else {
            usage.recordBlocked()
        }
    }

    fun usageSnapshot(deviceId: String, windowStart: Long, windowEnd: Long) =
        usage.snapshot(deviceId, windowStart, windowEnd)

    override fun onDestroy() {
        tunnel?.close()
        tunnel = null
        usage.reset()
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
                NotificationChannel(CHANNEL_ID, "FTN Family Guard", NotificationManager.IMPORTANCE_LOW)
            )
        }
    }

    companion object {
        private const val CHANNEL_ID = "ftn-family-guard"
        private const val NOTIFICATION_ID = 4102
    }
}
