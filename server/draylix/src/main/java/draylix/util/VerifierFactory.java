package draylix.util;

import draylix.proctocol.Verifier;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class VerifierFactory {
    private static Logger logger= LoggerFactory.getLogger(VerifierFactory.class);
    Verifier verifierImpl=null;
    public Verifier getVerifierImpl(){
        return verifierImpl;
    }
    public VerifierFactory(){}
    public void init(String className){
        if (className==null){
            logger.info("no token verification");
            return;
        }
        try {
            Class vc = Class.forName(className);
            Verifier verifier= (Verifier) vc.getConstructor().newInstance();
            verifierImpl=verifier;
            logger.info("using verifier : {}",className);
        } catch (Exception e) {
            logger.error("loading verifier failed : {}",e);
            logger.info("no token verification");
        }
    }
}
