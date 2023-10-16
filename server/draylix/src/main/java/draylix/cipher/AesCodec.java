package draylix.cipher;

import javax.crypto.BadPaddingException;
import javax.crypto.Cipher;
import javax.crypto.IllegalBlockSizeException;
import javax.crypto.NoSuchPaddingException;
import javax.crypto.spec.IvParameterSpec;
import javax.crypto.spec.SecretKeySpec;
import java.security.InvalidAlgorithmParameterException;
import java.security.InvalidKeyException;
import java.security.NoSuchAlgorithmException;

public class AesCodec {
    public static final int KeyLen=32;

    public static byte[] encrypt(byte[] data, byte[] key,byte[] iv) throws InvalidAlgorithmParameterException, InvalidKeyException, IllegalBlockSizeException, BadPaddingException {
        SecretKeySpec secretKeySpec = new SecretKeySpec(key, "AES");
        Cipher cipher = null;
        try {
            cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException(e);
        } catch (NoSuchPaddingException e) {
            throw new RuntimeException(e);
        }
        cipher.init(Cipher.ENCRYPT_MODE, secretKeySpec, new IvParameterSpec(iv));
        return cipher.doFinal(data);

    }

    public static byte[] decrypt(byte[] encryptedData, byte[] key,byte[] iv) throws InvalidAlgorithmParameterException, InvalidKeyException, IllegalBlockSizeException, BadPaddingException {
        SecretKeySpec secretKeySpec = new SecretKeySpec(key, "AES");
        Cipher cipher = null;
        try {
            cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException(e);
        } catch (NoSuchPaddingException e) {
            throw new RuntimeException(e);
        }
        cipher.init(Cipher.DECRYPT_MODE, secretKeySpec, new IvParameterSpec(iv));
        return cipher.doFinal(encryptedData);
    }

    //test
    public static void main(String[] args) throws Exception {
        byte[] sb="tongShangKuanYi".getBytes();
        byte[] keyb="12345678901234567890123456789012".getBytes();
        byte[] iv="HrGZo2uaSgDnL4Ke".getBytes();
        byte[] encrypt = encrypt(sb, keyb, iv);
        for (int i = 0; i <encrypt.length ; i++) {
            System.out.printf("%x ",encrypt[i]);
        }

        byte[] decrypt = decrypt(encrypt, keyb, iv);
        System.out.println(new String(decrypt));

    }
}
