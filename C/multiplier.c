/* All of our linux kernel includes. */
#include <linux/module.h>  /* Needed by all modules */
#include <linux/moduleparam.h>  /* Needed for module parameters */
#include <linux/kernel.h>  /* Needed for printk and KERN_* */
#include <linux/init.h>	   /* Need for __init macros  */
#include <linux/fs.h>	   /* Provides file ops structure */
#include <linux/sched.h>   /* Provides access to the "current" process
			      task structure */
#include <asm/uaccess.h>   /* Provides utilities to bring user space
			      data into kernel space.  Note, it is
			      processor arch specific. */
#include <asm/io.h> /* Needed for IO reads and writes */
#include "xparameters.h" /* needed for physical address of multiplier*/
#include <linux/slab.h> //kmalloc() and kfree()


/* Some defines */
#define DEVICE_NAME "multiplier"
#define BUF_LEN 80
/* From xparameters.h, physical address of multiplier */
#define PHY_ADDR XPAR_MULTIPLY_0_S00_AXI_BASEADDR 
/* Size of physical address range for multiply */
#define MEMSIZE XPAR_MULTIPLY_0_S00_AXI_HIGHADDR - XPAR_MULTIPLY_0_S00_AXI_BASEADDR + 1

void* virt_addr; // virtual address pointing to multiplier

/* Function prototypes, so we can setup the function pointers for dev
   file access correctly. */
int init_module(void);
void cleanup_module(void);
static int device_open(struct inode *, struct file *);
static int device_release(struct inode *, struct file *);
static ssize_t device_read(struct file *, char *, size_t, loff_t *);
static ssize_t device_write(struct file *, const char *, size_t, loff_t *);
static char *msg_bf_Ptr;	/* This time we'll use kmalloc and
				   kfree to handle the memory */

/* 
 * Global variables are declared as static, so are global but only
 * accessible within the file.
 */
static int Major;		/* Major number assigned to our device
				   driver */
static int Device_Open = 0;	/* Flag to signify open device */

/* This structure defines the function pointers to our functions for
   opening, closing, reading and writing the device file.  There are
   lots of other pointers in this structure which we are not using,
   see the whole definition in linux/fs.h */
static struct file_operations fops = {
  .read = device_read,
  .write = device_write,
  .open = device_open,
  .release = device_release
};

/* This function is run upon module load. This is where you setup data
   structures and reserve resources used by the module */
static int __init my_init(void)
{
    // Linux kernel's version of printf
    printk(KERN_INFO "Mapping virtual address...\n");

    // map virtual address to multiplier physical address using ioremap
    virt_addr = ioremap(PHY_ADDR, MEMSIZE);
		
    /* This function call registers a device and returns a major number
    associated with it.  Be wary, the device file could be accessed
    as soon as you register it, make sure anything you need (ie
    buffers ect) are setup _BEFORE_ you register the device.*/
    Major = register_chrdev(0, DEVICE_NAME, &fops);


    /* Negative values indicate a problem */
    if (Major < 0) {		
        /* Make sure you release any other resources you've already
        grabbed if you get here so you don't leave the kernel in a
        broken state. */
        printk(KERN_ALERT "Registering char device failed with %d\n", Major);

        /* We won't need our memory so make sure to free it here... */
        kfree(msg_bf_Ptr); 

        return Major;
    }
    // Print the major number to the kernel message buffer exactly as done in the examples provided.
    printk(KERN_INFO "Registered a device with dynamic Major number of %d\n", Major);
    printk(KERN_INFO "Create a device file for this device with this command:\n'mknod /dev/%s c %d 0'.\n", DEVICE_NAME, Major);
    return 0;		/* success */
}

/* This function is run just prior to the module's removal from the system.
You should release ALL resources used by your module here (otherwise be
prepared for a reboot). */
//From multiply.c
static void __exit my_exit(void)
{	
    //In the exit routine, unregister the device driver before the virtual memory unmapping.
    unregister_chrdev(Major, DEVICE_NAME);
      kfree(msg_bf_Ptr);		/* free our memory */
    printk(KERN_ALERT "unmapping virtual address space....\n");
    iounmap((void*)virt_addr);
}

/*For the open and close functions, do nothing except print to the kernel message buffer
informing the user when the device is opened and closed.*/

/* 
 * Called when a process tries to open the device file, like "cat
 * /dev/my_chardev".  Link to this function placed in file operations
 * structure for our device file.
 */
static int device_open(struct inode *inode, struct file *file)
{
    //Print that the device is open
    printk(KERN_ALERT "Device is now OPEN....\n");
    return 0;
}
//and closed
static int device_release(struct inode *inode, struct file *file)
{
    printk(KERN_ALERT "Device is now CLOSED\n");
    
    return 0;
}

/* 
 * Called when a process, which already opened the dev file, attempts
 * to read from it.
 */
static ssize_t device_read(struct file *filp, /* see include/linux/fs.h*/
			   char *buffer,      /* buffer to fill with data */
			   size_t length,     /* length of the buffer  */
			   loff_t * offset)
{
    /*
    * Number of bytes actually written to the buffer
    */
    int bytes_read = 0;
    //We will need a buffer similar to the following line from my_chardev_mem.c
    //but as an int
    //   msg_bf_Ptr = (char *)kmalloc(BUF_LEN*sizeof(char), GFP_KERNEL);
    int* readBuffer = (int*)kmalloc(length*sizeof(int), GFP_KERNEL);
    //need to allocate the buffer,
    //from multiply.c i know we need one for (virt_addr+0), (virt_addr+4), and (virt_addr+8) sinc its 0 through 12
	readBuffer[0] = ioread32(virt_addr);  // 0-3
	readBuffer[1] = ioread32(virt_addr + 4); //4-7
	readBuffer[2] = ioread32(virt_addr + 8);//8-11
    //now that its allocated we can actually read the chars
	char* charBuffer = (char*)readBuffer; 

    //make a for loop that goes through the buffer until it reaches thje end (length)
    int i; //'for' loop initial declarations are only allowed in C99 or C11 mode
    //I didnt know that smile
    for (i = 0; i < length; i++) { 
        //go through and put_user for each char of the buffer
        put_user(*(charBuffer++), buffer++); 
        //keep track of bytes_read
        bytes_read++;
    }
    //free memory 
    kfree(charBuffer);
    /* 
    * Most read functions return the number of bytes put into the
    * buffer
    */
    return bytes_read;
}

/* 
 * This function is called when somebody tries to write into our
 * device file.
 */
static ssize_t device_write(struct file *file, const char __user * buffer, size_t length, loff_t * offset)
{
  int i;
  //from exmple code
  msg_bf_Ptr = (char *)kmalloc(BUF_LEN*sizeof(char), GFP_KERNEL);
  /* printk(KERN_INFO "device_write(%p,%s,%d)", file, buffer, (int)length); */
  
  /* get_user pulls message from userspace into kernel space */
  //removed buff_length - 1 not necessart
  for (i = 0; i < length; i++)
  //increment buffer
    get_user(msg_bf_Ptr[i], buffer++);
  
  /* left one char early from buffer to leave space for null char*/
  msg_bf_Ptr[i] = '\0';

    //Now we have to write to registers :)
    int* regBuffer = (int*)msg_bf_Ptr; 
	
    //REMINDER: valid memory 0 - 7
	//register 0
	printk(KERN_INFO "Writing %d to register 0\n", regBuffer[0]);
	iowrite32(regBuffer[0], virt_addr+0); //Will cover 0-3
	
	//register 1
	printk(KERN_INFO "Writing %d to register 1\n", regBuffer[1]);
	iowrite32(regBuffer[1], virt_addr+4); //will cover 4-7
	
    //free memory of reg buffer since it was a pointer
	kfree(regBuffer);
  /* 
   * Again, return the number of input characters used 
   */
  return i;
}




/* These define info that can be displayed by modinfo */
MODULE_LICENSE("GPL");
MODULE_AUTHOR("Michael Mirhosseini (and others)"); //A large amount of this file was provided to us as a form of 'starter code'
MODULE_DESCRIPTION("Module which creates a character device and allows user interaction with it");

/* Here we define which functions we want to use for initialization and cleanup */
module_init(my_init);
module_exit(my_exit);
