#ifndef HEADER_SM_ERR_H
#define HEADER_SM_ERR_H

#include <syslog.h>

#ifdef  __cplusplus
extern "C" {
#endif

#define SM_RESOURCE_FREE(ptr, handler) \
    if(ptr) \
    { \
    handler(ptr); \
    ptr = NULL; \
}

#define SM_ERROR_ESCAPE(exp, errfunc, errmsg, errcode) \
    if (exp) \
    { \
    char errcode_str[512] = {0}; \
    BIO_snprintf(errcode_str,sizeof errcode_str,"[%s(%d)]    --Failed:(%04x%04x) [Reason:%s]", \
    __FILE__, __LINE__, (int)errfunc, (int)errmsg, ERR_error_string(ERR_peek_last_error(), NULL)); \
    syslog(LOG_USER|LOG_ERR, "%s\n", errcode_str); \
    goto _err; \
}

/* Function codes. */
#define SM_F_LOAD_SM2_PRIV_KEY_FROM_FILE  100
#define SM_F_LOAD_SM2_CERT_FROM_FILE      101
#define SM_F_SM2_NEW_ECKEY			      102
#define SM_F_SM2_FREE_ECKEY				  103
#define SM_F_SM2_SIGN			          104
#define SM_F_SM2_SIGN_DIRECT		      105
#define SM_F_SM2_VERIFY				      106
#define SM_F_SM2_VERIFY_DIRECT			  107
#define SM_F_SM3STARTS	                  108
#define SM_F_SM3UPDATE	                  109
#define SM_F_SM3FINISH	                  110
#define SM_F_SM4_SETKEY_ENC	              111
#define SM_F_SM4_SETKEY_DEC	              112
#define SM_F_SM4_CRYPT_ECB	              113
#define SM_F_SM4_CRYPT_CBC	              114
#define SM_F_SM4_ECB_ENCRYPT              115
#define SM_F_SM2_GEN_KEY                  116
#define SM_F_SM2_SIGN_SETUP               117
#define SM_F_SM2_DO_SIGN                  118
#define SM_F_SM2_DO_VERIFY                119
#define SM_F_SM2_SIGN_EX                  120
#define SM_F_LOAD_SM2_PRIV_KEY_FROM_BYTES 121
#define SM_F_LOAD_SM2_PUB_KEY_FROM_BYTES  122

/* Reason codes. */
#define SM_R_INVALID_PARAMETERS                            101
#define SM_R_BIO_NEW_FAILED                                102
#define SM_R_BIO_READ_FILENAME_FAILED                      103
#define SM_R_BN_BIN2BN_FAILED                              104
#define SM_R_BN_NEW_FAILED                                 105
#define SM_R_D2I_ECDSA_SIG_FAILED                          106
#define SM_R_EC_KEY_GENERATE_KEY_FAILED                    107
#define SM_R_EC_KEY_GET0_PRIVATE_KEY_FAILED                108
#define SM_R_EC_KEY_GET0_PUBLIC_KEY_FAILED                 109
#define SM_R_EC_KEY_NEW_BY_CURVE_NAME_FAILED               110
#define SM_R_EC_POINT_GET_AFFINE_COORDINATES_GFP_FAILED    111
#define SM_R_EVP_PKEY_GET1_EC_KEY_FAILED                   112
#define SM_R_PEM_READ_BIO_PRIVATEKEY_FAILED                113
#define SM_R_PEM_READ_BIO_X509_AUX_FAILED                  114
#define SM_R_SM2_GETBNBYTES_FAILED                         115
#define SM_R_SM2_SIGN_FAILED                               116
#define SM_R_SM2_VERIFY_FAILED                             117
#define SM_R_I2D_ECDSA_SIG_FAILED                          118
#define SM_R_ECDSA_SIG_NEW_FAILED                          119
#define SM_R_SM4_EXTENDKEY_FAILED                          120
#define SM_R_SM4_INVERTROUNDKEY_FAILED                     121
#define SM_R_SM4_CIPHERPLAINTEXTTRANSFORM_FAILED           122
#define SM_R_SM4_ECB_ENCRYPT_FAILED                        123
#define SM_R_SM4_CBC_ENCRYPT_FAILED                        124
#define SM_R_SM4_CBC_CHECK_PADDING_FAILED                  125
#define SM_R_EC_GROUP_NEW_FAILED                           126
#define SM_R_BN_HEX2BN_FAILED                              127
#define SM_R_EC_GROUP_SET_CURVE_GFP_FAILED                 128
#define SM_R_BN_CTX_NEW_FAILED                             129
#define SM_R_EC_POINT_NEW_FAILED                           130
#define SM_R_EC_POINT_SET_AFFINE_COORDINATES_GFP_FAILED    131
#define SM_R_EC_POINT_IS_ON_CURVE_FAILED                   132
#define SM_R_EC_GROUP_SET_GENERATOR_FAILED                 133
#define SM_R_EC_KEY_NEW_FAILED                             134
#define SM_R_EC_KEY_SET_GROUP_FAILED                       135
#define SM_R_EC_KEY_GET0_GROUP_FAILED                      136
#define SM_R_BN_RAND_RANGE_FAILED                          137
#define SM_R_EC_POINT_MUL_FAILED                           138
#define SM_R_EC_POINT_GET_AFFINE_COORDINATES_GF2M_FAILED   139
#define SM_R_BN_NNMOD_FAILED                               140
#define SM_R_EC_GROUP_GET_ORDER_FAILED                     141
#define SM_R_SM2_SIGN_SETUP_FAILED                         142
#define SM_R_BN_COPY_FAILED                                143
#define SM_R_BN_MOD_ADD_FAILED                             144
#define SM_R_NEED_NEW_SETUP_VALUES                         145
#define SM_R_BN_ONE_FAILED                                 146
#define SM_R_BN_MOD_INVERSE_FAILED                         147
#define SM_R_BN_MOD_MUL_FAILED                             148
#define SM_R_BN_MOD_SUB_FAILED                             149
#define SM_R_BAD_SIGNATURE                                 150
#define SM_R_BN_IS_ZERO_FAILED                             151
#define SM_R_SM2_DO_SIGN_EX_FAILED                         152
#define SM_R_SM2_DO_VERIFY_FAILED                          153
#define SM_R_BIO_NEW_MEM_FAILED                            154
#define SM_R_D2I_EC_PUBKEY_FAILED                          155
#ifdef  __cplusplus
}
#endif
#endif
