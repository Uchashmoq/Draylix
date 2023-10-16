package draylix;

import draylix.server.Server;
import draylix.util.Security;
import draylix.util.VerifierFactory;

import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.InputStream;
import java.util.Base64;
import java.util.Properties;

public class Main {
    public static void main(String[] args) {
        Properties cfg=readConfig();
        Server server = initServer(cfg);
        Security.REQ_DELAY=Long.parseLong(cfg.getProperty("validTime"));
        server.bind(cfg.getProperty("listenAddress"));
    }

    private static Server initServer(Properties cfg) {
        byte[] iv = decodeBase64(cfg.getProperty("initialVectorBase64"));
        byte[] ikey = decodeBase64(cfg.getProperty("initialKeyBase64"));
        Server server = new Server(iv, ikey);

        VerifierFactory verifierFactory=new VerifierFactory();
        verifierFactory.init(cfg.getProperty("tokenVerifierClass"));
        server.verifierFactory=verifierFactory;

        return server;
    }



    private static byte[] decodeBase64(String b64) {
        byte[] b;
        try {
            b = Base64.getDecoder().decode(b64);
        }catch (Exception e){
            throw new RuntimeException(b64, e);
        }
        return b;
    }

    private static Properties readConfig() {
        try(InputStream fis = new FileInputStream("configs/config.properties")) {
            if(fis==null) throw new FileNotFoundException();
            Properties cfg=new Properties();
            cfg.load(fis);
            return cfg;
        } catch (FileNotFoundException e) {
            throw new RuntimeException(e);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

}
