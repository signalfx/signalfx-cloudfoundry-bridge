package com.signalfx.cloudfoundry;

import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManager;
import javax.net.ssl.TrustManagerFactory;
import java.io.ByteArrayInputStream;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.FileSystems;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.security.KeyManagementException;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.cert.Certificate;
import java.security.cert.CertificateException;
import java.security.cert.CertificateFactory;

import static com.google.common.base.Preconditions.checkNotNull;

/**
 * SSL configuration
 */
public class SSL {
    /**
     * Configure SSL certificates
     *
     * @param conf
     * @throws IOException
     * @throws CertificateException
     * @throws KeyStoreException
     * @throws NoSuchAlgorithmException
     * @throws KeyManagementException
     */
    static void configureSsl(Configuration conf) throws IOException, CertificateException, KeyStoreException, NoSuchAlgorithmException, KeyManagementException {
        final String certString;

        switch (conf.getJmxSslCertSource()) {
            case FILE:
                certString = new String(Files.readAllBytes(FileSystems.getDefault().getPath(conf.getJmxSslCert())), "UTF-8");
                break;
            case INLINE:
                certString = conf.getJmxSslCert();
                break;
            default:
                throw new RuntimeException("Unable to load certificate from source");
        }

        CertificateFactory certificateFactory = CertificateFactory.getInstance("X.509");
        // Load default CA certs.
        KeyStore ks = KeyStore.getInstance(KeyStore.getDefaultType());
        try (FileInputStream stream = new FileInputStream(
                Paths.get(System.getProperty("java.home"), "lib/security/cacerts").toString())) {
            ks.load(stream, "changeit".toCharArray());
        }

        // Load the user-configured SSL cert.
        final Certificate certificate;
        try (InputStream is = new ByteArrayInputStream(certString.getBytes())) {
            certificate = certificateFactory.generateCertificate(is);
        }

        ks.setCertificateEntry("cloudfoundry", certificate);

        // Replace the keystore.
        // See http://stackoverflow.com/questions/32851341/load-ca-root-certificate-at-runtime-in-java
        TrustManagerFactory trustManagerFactory = TrustManagerFactory.getInstance("X509");
        trustManagerFactory.init(ks);
        SSLContext sslContext = SSLContext.getInstance("SSL");
        TrustManager[] tms = trustManagerFactory.getTrustManagers();
        sslContext.init(null, tms, null);
        SSLContext.setDefault(sslContext);
    }
}
