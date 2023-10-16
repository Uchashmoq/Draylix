package draylix.proctocol;

import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.InputStream;
import java.util.Properties;

public class PropertiesVerifier implements Verifier{
    public PropertiesVerifier(){
        tokens=new Properties();
        try(InputStream is = new FileInputStream("tokenWhiteList.properties")) {
            tokens.load(is);
        } catch (FileNotFoundException e) {
            throw new RuntimeException(e);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }
    private Properties tokens;
    @Override
    public boolean verifyToken(String token) {
        return tokens.containsKey(token);
    }

    public static void main(String[] args) {
        System.out.println(PropertiesVerifier.class.getName());
    }
}
