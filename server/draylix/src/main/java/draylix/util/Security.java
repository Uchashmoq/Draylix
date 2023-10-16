package draylix.util;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.security.SecureRandom;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public final class Security {
    private Security(){}
    public static int handShakeCheck=1;
    public static int maxAnno=8000;
    public static  long REQ_DELAY = 10;
    private static final Logger LOGGER= LoggerFactory.getLogger(Security.class);
    public static final Map<Long,Long> ANNOUNCES=new ConcurrentHashMap<>();

    private static boolean checkTime(long reqTime){
        long now = System.currentTimeMillis()/1000;
        if (Math.abs(now-reqTime)>REQ_DELAY){
            LOGGER.warn("now {},reqTime {},delay {} s",now,reqTime,now-reqTime);
            return false;
        }else{
            return true;
        }
    }

    private static  boolean addAnnounce (long announce,long timestamp) {
        if (ANNOUNCES.containsKey(announce)){
            return false;
        }
        if (ANNOUNCES.size()>maxAnno){
            ANNOUNCES.clear();
        }
        ANNOUNCES.put(announce,timestamp);
        return true;
    }

    public static byte[] generateRandomBytes(int length) {
        SecureRandom secureRandom = new SecureRandom();
        byte[] randomBytes = new byte[length];
        secureRandom.nextBytes(randomBytes);
        return randomBytes;
    }

    public static boolean checkTimeAnno(long timestamp,long anno){
        if(!checkTime(timestamp)) return  false;
        if(!addAnnounce(anno,timestamp)) return false;
        return true;
    }
}
